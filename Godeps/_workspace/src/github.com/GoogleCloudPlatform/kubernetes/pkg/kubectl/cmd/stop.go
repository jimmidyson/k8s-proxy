/*
Copyright 2014 Google Inc. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"fmt"
	"io"

	"github.com/GoogleCloudPlatform/kubernetes/pkg/kubectl/resource"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/util"
	"github.com/spf13/cobra"
)

func (f *Factory) NewCmdStop(out io.Writer) *cobra.Command {
	flags := &struct {
		Filenames util.StringList
	}{}
	cmd := &cobra.Command{
		Use:   "stop (<resource> <id>|-f filename)",
		Short: "Gracefully shut down a resource by id or filename.",
		Long: `Gracefully shut down a resource by id or filename.

Attempts to shut down and delete a resource that supports graceful termination.
If the resource is resizable it will be resized to 0 before deletion.

Examples:

    // Shut down foo.
    $ kubectl stop replicationcontroller foo

    // Shut down the service defined in service.json
    $ kubectl stop -f service.json

    // Shut down all resources in the path/to/resources directory
    $ kubectl stop -f path/to/resources`,
		Run: func(cmd *cobra.Command, args []string) {
			cmdNamespace, err := f.DefaultNamespace(cmd)
			checkErr(err)
			mapper, typer := f.Object(cmd)
			r := resource.NewBuilder(mapper, typer, f.ClientMapperForCommand(cmd)).
				ContinueOnError().
				NamespaceParam(cmdNamespace).RequireNamespace().
				ResourceTypeOrNameArgs(false, args...).
				FilenameParam(flags.Filenames...).
				Flatten().
				Do()
			checkErr(r.Err())

			r.Visit(func(info *resource.Info) error {
				reaper, err := f.Reaper(cmd, info.Mapping)
				checkErr(err)
				s, err := reaper.Stop(info.Namespace, info.Name)
				if err != nil {
					return err
				}
				fmt.Fprintf(out, "%s\n", s)
				return nil
			})
		},
	}
	cmd.Flags().VarP(&flags.Filenames, "filename", "f", "Filename, directory, or URL to file of resource(s) to be stopped")
	return cmd
}
