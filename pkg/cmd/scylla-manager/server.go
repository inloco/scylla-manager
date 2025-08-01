// Copyright (C) 2017 ScyllaDB

package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/scylladb/go-log"
	"github.com/scylladb/gocqlx/v2"
	config "github.com/scylladb/scylla-manager/v3/pkg/config/server"
	"github.com/scylladb/scylla-manager/v3/pkg/metrics"
	"github.com/scylladb/scylla-manager/v3/pkg/restapi"
	"github.com/scylladb/scylla-manager/v3/pkg/schema/table"
	"github.com/scylladb/scylla-manager/v3/pkg/service/backup"
	"github.com/scylladb/scylla-manager/v3/pkg/service/cluster"
	"github.com/scylladb/scylla-manager/v3/pkg/service/configcache"
	"github.com/scylladb/scylla-manager/v3/pkg/service/healthcheck"
	"github.com/scylladb/scylla-manager/v3/pkg/service/one2onerestore"
	"github.com/scylladb/scylla-manager/v3/pkg/service/repair"
	"github.com/scylladb/scylla-manager/v3/pkg/service/restore"
	"github.com/scylladb/scylla-manager/v3/pkg/service/scheduler"
	"github.com/scylladb/scylla-manager/v3/pkg/store"
	"github.com/scylladb/scylla-manager/v3/pkg/util/certutil"
	"github.com/scylladb/scylla-manager/v3/pkg/util/httppprof"
	"go.uber.org/multierr"
	"golang.org/x/sync/errgroup"
)

type server struct {
	config  config.Config
	session gocqlx.Session
	logger  log.Logger

	clusterSvc        *cluster.Service
	healthSvc         *healthcheck.Service
	backupSvc         *backup.Service
	restoreSvc        *restore.Service
	one2OneRestoreSvc one2onerestore.Servicer
	repairSvc         *repair.Service
	schedSvc          *scheduler.Service
	configCacheSvc    configcache.ConfigCacher

	httpServer       *http.Server
	httpsServer      *http.Server
	prometheusServer *http.Server
	debugServer      *http.Server

	errCh chan error
}

func newServer(c config.Config, logger log.Logger) (*server, error) {
	session, err := gocqlx.WrapSession(gocqlClusterConfig(c).CreateSession())
	if err != nil {
		return nil, errors.Wrapf(err, "database")
	}

	return &server{
		config:  c,
		session: session,
		logger:  logger,

		errCh: make(chan error, 4),
	}, nil
}

func (s *server) makeServices(ctx context.Context) error {
	var err error

	drawerStore := store.NewTableStore(s.session, table.Drawer)
	secretsStore := store.NewTableStore(s.session, table.Secrets)

	s.clusterSvc, err = cluster.NewService(s.session, metrics.NewClusterMetrics().MustRegister(), secretsStore, s.config.TimeoutConfig,
		s.config.ClientCacheTimeout, s.logger.Named("cluster"))
	if err != nil {
		return errors.Wrapf(err, "cluster service")
	}
	s.clusterSvc.SetOnChangeListener(s.onClusterChange)

	s.configCacheSvc = configcache.NewService(s.config.ConfigCache, s.clusterSvc, s.clusterSvc.CreateClientNoCache,
		secretsStore, s.logger)
	s.initConfigCacheSvc(ctx)

	s.healthSvc, err = healthcheck.NewService(
		s.config.Healthcheck,
		s.clusterSvc.Client,
		secretsStore,
		s.clusterSvc.GetClusterByID,
		s.configCacheSvc,
		s.logger.Named("healthcheck"),
	)
	if err != nil {
		return errors.Wrapf(err, "healthcheck service")
	}

	s.repairSvc, err = repair.NewService(
		s.session,
		s.config.Repair,
		metrics.NewRepairMetrics().MustRegister(),
		s.clusterSvc.Client,
		s.clusterSvc.GetSession,
		s.configCacheSvc,
		s.logger.Named("repair"),
	)
	if err != nil {
		return errors.Wrapf(err, "repair service")
	}

	s.backupSvc, err = backup.NewService(
		s.session,
		s.config.Backup,
		metrics.NewBackupMetrics().MustRegister(),
		s.clusterSvc.GetClusterName,
		s.clusterSvc.Client,
		s.clusterSvc.GetSession,
		s.configCacheSvc,
		s.logger.Named("backup"),
	)
	if err != nil {
		return errors.Wrapf(err, "backup service")
	}

	s.restoreSvc, err = restore.NewService(
		s.repairSvc,
		s.session,
		s.config.Restore,
		metrics.NewRestoreMetrics().MustRegister(),
		s.clusterSvc.Client,
		s.clusterSvc.GetSession,
		s.configCacheSvc,
		s.logger.Named("restore"),
	)
	if err != nil {
		return errors.Wrapf(err, "restore service")
	}

	s.one2OneRestoreSvc, err = one2onerestore.NewService(
		s.repairSvc,
		s.session,
		s.clusterSvc.Client,
		s.clusterSvc.GetSession,
		s.configCacheSvc,
		s.logger.Named("one2onerestore"),
		metrics.NewOne2OneRestoreMetrics().MustRegister(),
	)
	if err != nil {
		return errors.Wrapf(err, "one2onerestore service")
	}

	s.schedSvc, err = scheduler.NewService(
		s.session,
		metrics.NewSchedulerMetrics().MustRegister(),
		drawerStore,
		s.logger.Named("scheduler"),
	)
	if err != nil {
		return errors.Wrapf(err, "scheduler service")
	}

	restoreExclusiveLock := scheduler.NewTaskExclusiveLockPolicy(scheduler.RestoreTask, scheduler.One2OneRestoreTask)

	// Register the runners
	s.schedSvc.SetRunner(scheduler.BackupTask,
		scheduler.PolicyRunner{Policy: restoreExclusiveLock, Runner: s.backupSvc.Runner(), TaskType: scheduler.BackupTask})
	s.schedSvc.SetRunner(scheduler.RestoreTask,
		scheduler.PolicyRunner{Policy: restoreExclusiveLock, Runner: s.restoreSvc.Runner(), TaskType: scheduler.RestoreTask})
	s.schedSvc.SetRunner(scheduler.One2OneRestoreTask,
		scheduler.PolicyRunner{Policy: restoreExclusiveLock, Runner: s.one2OneRestoreSvc.Runner(), TaskType: scheduler.One2OneRestoreTask},
	)
	s.schedSvc.SetRunner(scheduler.HealthCheckTask, s.healthSvc.Runner())
	s.schedSvc.SetRunner(scheduler.RepairTask,
		scheduler.PolicyRunner{Policy: restoreExclusiveLock, Runner: s.repairSvc.Runner(), TaskType: scheduler.RepairTask})
	s.schedSvc.SetRunner(scheduler.ValidateBackupTask, s.backupSvc.ValidationRunner())

	// Add additional properties on task run.
	// This is a bit hacky way of providing selected information on other tasks
	// such as locations, retention and passing it in context of a task run.

	s.schedSvc.SetPropertiesDecorator(scheduler.BackupTask, s.backupSvc.TaskDecorator(s.schedSvc))
	s.schedSvc.SetPropertiesDecorator(scheduler.ValidateBackupTask, s.backupSvc.ValidateBackupTaskDecorator(s.schedSvc))

	return nil
}

func (s *server) onClusterChange(ctx context.Context, c cluster.Change) error {
	switch c.Type {
	case cluster.Update:
		go s.configCacheSvc.ForceUpdateCluster(context.Background(), c.ID)
	case cluster.Create:
		s.configCacheSvc.ForceUpdateCluster(ctx, c.ID)
		for _, t := range makeAutoHealthCheckTasks(c.ID, s.config.Healthcheck) {
			if err := s.schedSvc.PutTask(ctx, t); err != nil {
				return errors.Wrapf(err, "add automatically scheduled health check for cluster %s", c.ID)
			}
		}
		if !c.WithoutRepair {
			if err := s.schedSvc.PutTask(ctx, makeAutoRepairTask(c.ID)); err != nil {
				return errors.Wrapf(err, "add automatically scheduled weekly repair for cluster %s", c.ID)
			}
		}
	case cluster.Delete:
		s.configCacheSvc.RemoveCluster(c.ID)
		tasks, err := s.schedSvc.ListTasks(ctx, c.ID, scheduler.ListFilter{Disabled: true, Short: true})
		if err != nil {
			return errors.Wrapf(err, "find this cluster %s tasks", c.ID)
		}
		var errs error
		for _, t := range tasks {
			errs = multierr.Append(errs, s.schedSvc.DeleteTask(ctx, &t.Task))
		}
		if errs != nil {
			return errors.Wrapf(errs, "remove cluster %s tasks", c.ID)
		}
	}

	return nil
}

func (s *server) makeServers(ctx context.Context) error {
	services := restapi.Services{
		Cluster:        s.clusterSvc,
		HealthCheck:    s.healthSvc,
		Repair:         s.repairSvc,
		Backup:         s.backupSvc,
		Restore:        s.restoreSvc,
		Scheduler:      s.schedSvc,
		One2OneRestore: s.one2OneRestoreSvc,
	}
	h := restapi.New(services, s.logger.Named("http"))

	if s.config.HTTP != "" {
		s.httpServer = &http.Server{
			Addr:    s.config.HTTP,
			Handler: h,
		}
	}
	if s.config.HTTPS != "" {
		tlsConfig, err := s.tlsConfig(ctx)
		if err != nil {
			return errors.Wrapf(err, "tls")
		}
		s.httpsServer = &http.Server{
			Addr:      s.config.HTTPS,
			TLSConfig: tlsConfig,
			Handler:   h,
		}
	}
	if s.config.Prometheus != "" {
		s.prometheusServer = &http.Server{
			Addr:    s.config.Prometheus,
			Handler: restapi.NewPrometheus(s.clusterSvc),
		}
	}
	if s.config.Debug != "" {
		s.debugServer = &http.Server{
			Addr:    s.config.Debug,
			Handler: httppprof.Handler(),
		}
	}

	return nil
}

func (s *server) initConfigCacheSvc(ctx context.Context) {
	s.logger.Info(ctx, "Initializing the clusters configuration cache")
	defer s.logger.Info(ctx, "Clusters config cache initialized")

	s.configCacheSvc.Init(ctx)
}

func (s *server) startConfigCacheSvcAsync(ctx context.Context) {
	go func() {
		logger := s.logger.Named("Config cache goroutine")

		for {
			select {
			case <-ctx.Done():
				logger.Info(ctx, "Shutdown")
				return
			default:
				func() {
					defer func() {
						if r := recover(); r != nil {
							logger.Error(ctx, "Recovered from panic:", "recover", r)
						}

						s.configCacheSvc.Init(ctx)
					}()

					s.configCacheSvc.Run(ctx)
				}()
			}
		}
	}()
}

func (s *server) tlsConfig(ctx context.Context) (*tls.Config, error) {
	var (
		cert tls.Certificate
		err  error
	)
	if s.config.HasTLSCert() {
		s.logger.Info(ctx, "Loading TLS certificate from disk",
			"cert_file", s.config.TLSCertFile,
			"key_file", s.config.TLSKeyFile,
		)
		cert, err = tls.LoadX509KeyPair(s.config.TLSCertFile, s.config.TLSKeyFile)
	} else {
		hosts := []string{s.config.HTTPS}
		s.logger.Info(ctx, "Generating TLS certificate", "hosts", hosts)
		cert, err = certutil.GenerateSelfSignedCertificate(hosts)
	}
	if err != nil {
		return nil, errors.Wrap(err, "certificate")
	}
	tlsConfig := s.config.TLSVersion.TLSConfig()
	tlsConfig.Certificates = []tls.Certificate{cert}

	if s.config.TLSCAFile != "" {
		pool, err := s.certPool(s.config.TLSCAFile)
		if err != nil {
			return nil, errors.Wrap(err, "CA")
		}
		tlsConfig.ClientCAs = pool
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
	}

	return tlsConfig, nil
}

func (s *server) certPool(file string) (*x509.CertPool, error) {
	pem, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(pem) {
		return nil, errors.Errorf("no certificates found in %s", file)
	}
	return pool, nil
}

func (s *server) startServices(ctx context.Context) error {
	if err := s.clusterSvc.Init(ctx); err != nil {
		return errors.Wrap(err, "cluster service")
	}
	if err := s.schedSvc.UpdateHealthcheckTasks(ctx, s.config.Healthcheck); err != nil {
		return errors.Wrap(err, "update healthcheck tasks")
	}
	if err := s.schedSvc.LoadTasks(ctx); err != nil {
		return errors.Wrap(err, "schedule service")
	}

	s.startConfigCacheSvcAsync(ctx)

	return nil
}

func (s *server) startServers(ctx context.Context) {
	if s.httpServer != nil {
		s.logger.Info(ctx, "Starting HTTP server", "address", s.httpServer.Addr)
		go func() {
			s.errCh <- s.httpServer.ListenAndServe()
		}()
	}

	if s.httpsServer != nil {
		s.logger.Info(ctx, "Starting HTTPS server", "address", s.httpsServer.Addr, "client_ca", s.config.TLSCAFile)
		go func() {
			s.errCh <- errors.Wrap(s.httpsServer.ListenAndServeTLS("", ""), "HTTPS server start")
		}()
	}

	if s.prometheusServer != nil {
		s.logger.Info(ctx, "Starting Prometheus server", "address", s.prometheusServer.Addr)
		go func() {
			s.errCh <- errors.Wrap(s.prometheusServer.ListenAndServe(), "prometheus server start")
		}()
	}

	if s.debugServer != nil {
		s.logger.Info(ctx, "Starting debug server", "address", s.debugServer.Addr)
		go func() {
			s.errCh <- errors.Wrap(s.debugServer.ListenAndServe(), "debug server start")
		}()
	}

	s.logger.Info(ctx, "Service started")
}

func (s *server) shutdownServers(ctx context.Context, timeout time.Duration) {
	s.logger.Info(ctx, "Closing servers", "timeout", timeout)

	tctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var eg errgroup.Group
	eg.Go(s.shutdownHTTPServer(tctx, s.httpServer))
	eg.Go(s.shutdownHTTPServer(tctx, s.httpsServer))
	eg.Go(s.shutdownHTTPServer(tctx, s.prometheusServer))
	eg.Go(s.shutdownHTTPServer(tctx, s.debugServer))
	eg.Wait() // nolint: errcheck
}

func (s *server) shutdownHTTPServer(ctx context.Context, server *http.Server) func() error {
	return func() error {
		if server == nil {
			return nil
		}
		if err := server.Shutdown(ctx); err != nil {
			s.logger.Info(ctx, "Closing server failed", "address", server.Addr, "error", err)
		} else {
			s.logger.Info(ctx, "Closing server done", "address", server.Addr)
		}

		// Force close
		return server.Close()
	}
}

func (s *server) close() {
	// The cluster service needs to be closed last because it handles closing of
	// connections to agent running on the nodes.
	s.schedSvc.Close()
	s.clusterSvc.Close()

	s.session.Close()
}
