// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"net/http"
	"time"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	cr "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
)

// NewGetClusterClusterIDBackupsSchemaParams creates a new GetClusterClusterIDBackupsSchemaParams object
// with the default values initialized.
func NewGetClusterClusterIDBackupsSchemaParams() *GetClusterClusterIDBackupsSchemaParams {
	var ()
	return &GetClusterClusterIDBackupsSchemaParams{

		timeout: cr.DefaultTimeout,
	}
}

// NewGetClusterClusterIDBackupsSchemaParamsWithTimeout creates a new GetClusterClusterIDBackupsSchemaParams object
// with the default values initialized, and the ability to set a timeout on a request
func NewGetClusterClusterIDBackupsSchemaParamsWithTimeout(timeout time.Duration) *GetClusterClusterIDBackupsSchemaParams {
	var ()
	return &GetClusterClusterIDBackupsSchemaParams{

		timeout: timeout,
	}
}

// NewGetClusterClusterIDBackupsSchemaParamsWithContext creates a new GetClusterClusterIDBackupsSchemaParams object
// with the default values initialized, and the ability to set a context for a request
func NewGetClusterClusterIDBackupsSchemaParamsWithContext(ctx context.Context) *GetClusterClusterIDBackupsSchemaParams {
	var ()
	return &GetClusterClusterIDBackupsSchemaParams{

		Context: ctx,
	}
}

// NewGetClusterClusterIDBackupsSchemaParamsWithHTTPClient creates a new GetClusterClusterIDBackupsSchemaParams object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewGetClusterClusterIDBackupsSchemaParamsWithHTTPClient(client *http.Client) *GetClusterClusterIDBackupsSchemaParams {
	var ()
	return &GetClusterClusterIDBackupsSchemaParams{
		HTTPClient: client,
	}
}

/*
GetClusterClusterIDBackupsSchemaParams contains all the parameters to send to the API endpoint
for the get cluster cluster ID backups schema operation typically these are written to a http.Request
*/
type GetClusterClusterIDBackupsSchemaParams struct {

	/*ClusterID
	  ID of the cluster which will fetch backed up schema

	*/
	ClusterID string
	/*Location
	  Location of the backup which schema will be fetched. In case of a multi-location backup, any location can be specified

	*/
	Location string
	/*QueryClusterID
	  ID of the cluster which backed up schema will be fetched. Used for filtering purposes

	*/
	QueryClusterID *string
	/*QueryTaskID
	  ID of the task which backed up schema. Used for filtering purposes

	*/
	QueryTaskID *string
	/*SnapshotTag
	  Snapshot tag of the backup which schema will be fetched

	*/
	SnapshotTag string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the get cluster cluster ID backups schema params
func (o *GetClusterClusterIDBackupsSchemaParams) WithTimeout(timeout time.Duration) *GetClusterClusterIDBackupsSchemaParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the get cluster cluster ID backups schema params
func (o *GetClusterClusterIDBackupsSchemaParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the get cluster cluster ID backups schema params
func (o *GetClusterClusterIDBackupsSchemaParams) WithContext(ctx context.Context) *GetClusterClusterIDBackupsSchemaParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the get cluster cluster ID backups schema params
func (o *GetClusterClusterIDBackupsSchemaParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the get cluster cluster ID backups schema params
func (o *GetClusterClusterIDBackupsSchemaParams) WithHTTPClient(client *http.Client) *GetClusterClusterIDBackupsSchemaParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the get cluster cluster ID backups schema params
func (o *GetClusterClusterIDBackupsSchemaParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithClusterID adds the clusterID to the get cluster cluster ID backups schema params
func (o *GetClusterClusterIDBackupsSchemaParams) WithClusterID(clusterID string) *GetClusterClusterIDBackupsSchemaParams {
	o.SetClusterID(clusterID)
	return o
}

// SetClusterID adds the clusterId to the get cluster cluster ID backups schema params
func (o *GetClusterClusterIDBackupsSchemaParams) SetClusterID(clusterID string) {
	o.ClusterID = clusterID
}

// WithLocation adds the location to the get cluster cluster ID backups schema params
func (o *GetClusterClusterIDBackupsSchemaParams) WithLocation(location string) *GetClusterClusterIDBackupsSchemaParams {
	o.SetLocation(location)
	return o
}

// SetLocation adds the location to the get cluster cluster ID backups schema params
func (o *GetClusterClusterIDBackupsSchemaParams) SetLocation(location string) {
	o.Location = location
}

// WithQueryClusterID adds the queryClusterID to the get cluster cluster ID backups schema params
func (o *GetClusterClusterIDBackupsSchemaParams) WithQueryClusterID(queryClusterID *string) *GetClusterClusterIDBackupsSchemaParams {
	o.SetQueryClusterID(queryClusterID)
	return o
}

// SetQueryClusterID adds the queryClusterId to the get cluster cluster ID backups schema params
func (o *GetClusterClusterIDBackupsSchemaParams) SetQueryClusterID(queryClusterID *string) {
	o.QueryClusterID = queryClusterID
}

// WithQueryTaskID adds the queryTaskID to the get cluster cluster ID backups schema params
func (o *GetClusterClusterIDBackupsSchemaParams) WithQueryTaskID(queryTaskID *string) *GetClusterClusterIDBackupsSchemaParams {
	o.SetQueryTaskID(queryTaskID)
	return o
}

// SetQueryTaskID adds the queryTaskId to the get cluster cluster ID backups schema params
func (o *GetClusterClusterIDBackupsSchemaParams) SetQueryTaskID(queryTaskID *string) {
	o.QueryTaskID = queryTaskID
}

// WithSnapshotTag adds the snapshotTag to the get cluster cluster ID backups schema params
func (o *GetClusterClusterIDBackupsSchemaParams) WithSnapshotTag(snapshotTag string) *GetClusterClusterIDBackupsSchemaParams {
	o.SetSnapshotTag(snapshotTag)
	return o
}

// SetSnapshotTag adds the snapshotTag to the get cluster cluster ID backups schema params
func (o *GetClusterClusterIDBackupsSchemaParams) SetSnapshotTag(snapshotTag string) {
	o.SnapshotTag = snapshotTag
}

// WriteToRequest writes these params to a swagger request
func (o *GetClusterClusterIDBackupsSchemaParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	// path param cluster_id
	if err := r.SetPathParam("cluster_id", o.ClusterID); err != nil {
		return err
	}

	// query param location
	qrLocation := o.Location
	qLocation := qrLocation
	if qLocation != "" {
		if err := r.SetQueryParam("location", qLocation); err != nil {
			return err
		}
	}

	if o.QueryClusterID != nil {

		// query param query_cluster_id
		var qrQueryClusterID string
		if o.QueryClusterID != nil {
			qrQueryClusterID = *o.QueryClusterID
		}
		qQueryClusterID := qrQueryClusterID
		if qQueryClusterID != "" {
			if err := r.SetQueryParam("query_cluster_id", qQueryClusterID); err != nil {
				return err
			}
		}

	}

	if o.QueryTaskID != nil {

		// query param query_task_id
		var qrQueryTaskID string
		if o.QueryTaskID != nil {
			qrQueryTaskID = *o.QueryTaskID
		}
		qQueryTaskID := qrQueryTaskID
		if qQueryTaskID != "" {
			if err := r.SetQueryParam("query_task_id", qQueryTaskID); err != nil {
				return err
			}
		}

	}

	// query param snapshot_tag
	qrSnapshotTag := o.SnapshotTag
	qSnapshotTag := qrSnapshotTag
	if qSnapshotTag != "" {
		if err := r.SetQueryParam("snapshot_tag", qSnapshotTag); err != nil {
			return err
		}
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
