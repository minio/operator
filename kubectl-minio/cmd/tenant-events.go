// This file is part of MinIO Operator
// Copyright (C) 2022, MinIO, Inc.
//
// This code is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License, version 3,
// as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License, version 3,
// along with this program.  If not, see <http://www.gnu.org/licenses/>

package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/minio/kubectl-minio/cmd/helpers"
	v2 "github.com/minio/operator/pkg/apis/minio.min.io/v2"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/duration"
	"k8s.io/klog/v2"
	"sigs.k8s.io/yaml"
)

const (
	eventsDesc = `'events' command displays tenant events`
)

type eventsCmd struct {
	out        io.Writer
	errOut     io.Writer
	ns         string
	yamlOutput bool
	jsonOutput bool
}

func newTenantEventsCmd(out io.Writer, errOut io.Writer) *cobra.Command {
	c := &eventsCmd{out: out, errOut: errOut}

	cmd := &cobra.Command{
		Use:     "events <TENANTNAME>",
		Short:   "Display tenant events",
		Long:    eventsDesc,
		Example: `  kubectl minio events tenant1`,
		Args: func(cmd *cobra.Command, args []string) error {
			return c.validate(args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			err := c.run(args)
			if err != nil {
				klog.Warning(err)
				return err
			}
			return nil
		},
	}
	cmd = helpers.DisableHelp(cmd)
	f := cmd.Flags()
	f.StringVarP(&c.ns, "namespace", "n", "", "namespace scope for this request")
	f.BoolVarP(&c.yamlOutput, "yaml", "y", false, "yaml output")
	f.BoolVarP(&c.jsonOutput, "json", "j", false, "json output")
	return cmd
}

func (d *eventsCmd) validate(args []string) error {
	return validateTenantArgs("events", args)
}

func (d *eventsCmd) run(args []string) error {
	path, _ := rootCmd.Flags().GetString(kubeconfig)
	oclient, err := helpers.GetKubeOperatorClient(path)
	if err != nil {
		return err
	}
	if d.ns == "" || d.ns == helpers.DefaultNamespace {
		d.ns, err = getTenantNamespace(oclient, args[0])
		if err != nil {
			return err
		}
	}
	tenant, err := oclient.MinioV2().Tenants(d.ns).Get(context.Background(), args[0], metav1.GetOptions{})
	if err != nil {
		return err
	}
	return d.printTenantEvents(tenant)
}

func (d *eventsCmd) printTenantEvents(tenant *v2.Tenant) error {
	events, err := d.getTenantEvents(tenant)
	if err != nil {
		return err
	}
	if !d.jsonOutput && !d.yamlOutput {
		d.printRawTenantEvents(events)
		return nil
	}
	if d.jsonOutput && d.yamlOutput {
		return fmt.Errorf("Only one output can be used to display events")
	}
	if d.jsonOutput {
		return d.printJSONTenantEvents(events)
	}
	if d.yamlOutput {
		return d.printYAMLTenantEvents(events)
	}
	return nil
}

func (d *eventsCmd) getTenantEvents(tenant *v2.Tenant) (*v1.EventList, error) {
	path, _ := rootCmd.Flags().GetString(kubeconfig)
	kubeClient, err := helpers.GetKubeClient(path)
	if err != nil {
		return nil, err
	}
	return kubeClient.CoreV1().Events(tenant.Namespace).List(context.Background(), metav1.ListOptions{})
}

func (d *eventsCmd) printRawTenantEvents(events *v1.EventList) {
	table := d.initEventsTable()
	data := d.getEventsData(events)
	table.AppendBulk(data)
	table.Render()
}

func (d *eventsCmd) initEventsTable() *tablewriter.Table {
	table := tablewriter.NewWriter(d.out)
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("\t")
	table.SetNoWhiteSpace(true)
	table.SetHeader([]string{"LAST SEEN", "TYPE", "REASON", "OBJECT", "MESSAGE"})
	return table
}

func (d *eventsCmd) getEventsData(events *v1.EventList) (data [][]string) {
	for _, event := range events.Items {
		data = append(data, []string{
			duration.HumanDuration(time.Since(event.LastTimestamp.Time)),
			event.Type,
			event.Reason,
			event.InvolvedObject.Name,
			event.Message,
		})
	}
	return data
}

func (d *eventsCmd) printJSONTenantEvents(events *v1.EventList) error {
	enc := json.NewEncoder(d.out)
	enc.SetIndent("", " ")
	return enc.Encode(events)
}

func (d *eventsCmd) printYAMLTenantEvents(events *v1.EventList) error {
	eventsYAML, err := yaml.Marshal(events)
	if err != nil {
		return err
	}
	fmt.Fprintln(d.out, string(eventsYAML))
	return nil
}
