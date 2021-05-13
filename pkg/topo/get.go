// Copyright 2019-present Open Networking Foundation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package topo

import (
	"bytes"
	"context"
	"fmt"
	topoapi "github.com/onosproject/onos-api/go/onos/topo"
	"github.com/onosproject/onos-lib-go/pkg/cli"
	"github.com/spf13/cobra"
	"io"
	"text/tabwriter"
	"time"
)

func getGetEntityCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "entity <id>",
		Aliases: []string{"entities"},
		Args:    cobra.MaximumNArgs(1),
		Short:   "Get Entity",
		RunE:    runGetEntityCommand,
	}
	cmd.Flags().Bool("no-headers", false, "disables output headers")
	cmd.Flags().BoolP("verbose", "v", false, "verbose output")
	return cmd
}

func getGetRelationCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "relation <id>",
		Aliases: []string{"relations"},
		Args:    cobra.MaximumNArgs(1),
		Short:   "Get Relation",
		RunE:    runGetRelationCommand,
	}
	cmd.Flags().Bool("no-headers", false, "disables output headers")
	cmd.Flags().BoolP("verbose", "v", false, "verbose output")
	return cmd
}

func getGetKindCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "kind <id>",
		Aliases: []string{"kinds"},
		Args:    cobra.MaximumNArgs(1),
		Short:   "Get Kind",
		RunE:    runGetKindCommand,
	}
	cmd.Flags().Bool("no-headers", false, "disables output headers")
	cmd.Flags().BoolP("verbose", "v", false, "verbose output")
	return cmd
}

func runGetEntityCommand(cmd *cobra.Command, args []string) error {
	return runGetCommand(cmd, args, topoapi.Object_ENTITY)
}

func runGetRelationCommand(cmd *cobra.Command, args []string) error {
	return runGetCommand(cmd, args, topoapi.Object_RELATION)
}

func runGetKindCommand(cmd *cobra.Command, args []string) error {
	return runGetCommand(cmd, args, topoapi.Object_KIND)
}

func runGetCommand(cmd *cobra.Command, args []string, objectType topoapi.Object_Type) error {
	noHeaders, _ := cmd.Flags().GetBool("no-headers")
	verbose, _ := cmd.Flags().GetBool("verbose")

	writer := new(tabwriter.Writer)
	writer.Init(cli.GetOutput(), 0, 0, 3, ' ', tabwriter.FilterHTML)

	if !noHeaders {
		printHeader(writer, objectType, verbose, false)
	}

	if len(args) == 0 {
		objects, err := listObjects(cmd)
		if err == nil {
			for _, object := range objects {
				if objectType == object.Type {
					printObject(writer, object, verbose)
				}
			}
		}
	} else {
		id := args[0]
		object, err := getObject(cmd, topoapi.ID(id))
		if err != nil {
			return err
		}
		if object != nil && objectType == object.Type {
			printObject(writer, *object, verbose)
		}
	}

	return nil
}

func listObjects(cmd *cobra.Command) ([]topoapi.Object, error) {
	conn, err := cli.GetConnection(cmd)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := topoapi.CreateTopoClient(conn)

	resp, err := client.List(context.Background(), &topoapi.ListRequest{})
	if err != nil {
		return nil, err
	}
	return resp.Objects, nil
}

func getObject(cmd *cobra.Command, id topoapi.ID) (*topoapi.Object, error) {
	conn, err := cli.GetConnection(cmd)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := topoapi.CreateTopoClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	response, err := client.Get(ctx, &topoapi.GetRequest{ID: id})
	if err != nil {
		cli.Output("get error")
		return nil, err
	}
	return response.Object, nil
}

func printHeader(writer io.Writer, objectType topoapi.Object_Type, verbose bool, printUpdateType bool) {
	if printUpdateType {
		_, _ = fmt.Fprintf(writer, "%-*.*s", width, prec, "Update Type")
	}

	if objectType == topoapi.Object_ENTITY {
		_, _ = fmt.Fprintf(writer, "%-*.*s%-*.*s\n%-*.*s", width, prec, "Object Type",
			width, prec, "Entity ID", width, prec, "Kind ID")
	} else if objectType == topoapi.Object_RELATION {
		_, _ = fmt.Fprintf(writer, "%-*.*s%-*.*s\n%-*.*s%-*.*s%-*.*s", width, prec, "Object Type",
			width, prec, "Relation ID", width, prec, "Kind ID", width, prec, "Source ID", width, prec, "Target ID")
	} else if objectType == topoapi.Object_KIND {
		_, _ = fmt.Fprintf(writer, "%-*.*s%-*.*s\n%-*.*s", width, prec, "Object Type",
			width, prec, "Kind ID", width, prec, "Name")
	}

	if !verbose {
		_, _ = fmt.Fprintf(writer, "\tAspects\n")
	}
}

const (
	width = 16
	prec  = width - 1
)

func printObject(writer io.Writer, object topoapi.Object, verbose bool) {
	switch object.Type {
	case topoapi.Object_ENTITY:
		var kindID topoapi.ID
		if e := object.GetEntity(); e != nil {
			kindID = e.KindID
		}
		_, _ = fmt.Fprintf(writer, "%-*.*s%-*.*s%-*.*s", width, prec, object.Type, width, prec, object.ID, width, prec, kindID)
		printAspects(writer, object, verbose)

	case topoapi.Object_RELATION:
		r := object.GetRelation()
		_, _ = fmt.Fprintf(writer, "%-*.*s%-*.*s%-*.*s%-*.*s%-*.*s", width, prec, object.Type, width, prec, object.ID, width, prec, r.KindID,
			width, prec, r.SrcEntityID, width, prec, r.TgtEntityID)
		printAspects(writer, object, verbose)

	case topoapi.Object_KIND:
		k := object.GetKind()
		_, _ = fmt.Fprintf(writer, "%-*.*s%-*.*s%-*.*s", width, prec, object.Type, width, prec, object.ID, width, prec, k.GetName())
		printAspects(writer, object, verbose)

	default:
		_, _ = fmt.Fprintf(writer, "\n")
	}
}

func printAspects(writer io.Writer, object topoapi.Object, verbose bool) {
	if verbose {
		for aspectType, aspect := range object.Aspects {
			_, _ = fmt.Fprintf(writer, "\t%s=%s\n", aspectType, bytes.NewBuffer(aspect.Value).String())
		}
	} else {
		_, _ = fmt.Fprintf(writer, "\t%s\n", aspectList(object))
	}
}

func aspectList(object topoapi.Object) string {
	buf := bytes.Buffer{}
	first := true
	if object.Aspects != nil {
		for aspectType := range object.Aspects {
			if !first {
				buf.WriteString(",")
			} else {
				first = false
			}
			buf.WriteString(aspectType)
		}
	}
	return buf.String()
}
