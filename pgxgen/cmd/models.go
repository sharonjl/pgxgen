// Copyright © 2018 Sharon Lourduraj
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

package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/jackc/pgx"
	"github.com/reiver/go-stringcase"
	"github.com/sharonjl/pgxgen"
	"github.com/spf13/cobra"
)

// modelsCmd represents the models command
var modelsCmd = &cobra.Command{
	Use:   "models",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: modelRunFn,
}

func init() {
	rootCmd.AddCommand(modelsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// modelsCmd.PersistentFlags().String("foo", "", "A help for foo")
	modelsCmd.PersistentFlags().String("package", "", "output package name")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// modelsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func modelRunFn(cmd *cobra.Command, args []string) {
	// Output directory
	outf := cmd.Flag("out").Value.String()
	outdir, err := filepath.Abs(outf)
	if err != nil {
		panic("output directory: " + outf + ": " + err.Error())
	}

	// Package name
	pkgName := stringcase.ToCamelCase(filepath.Base(outdir))
	if cmd.Flag("package").Changed {
		pkgName = cmd.Flag("package").Value.String()
	}

	conn, err := pgx.Connect(pgx.ConnConfig{
		Host:     dbHost,
		User:     dbUser,
		Password: dbPassword,
		Database: dbName,
	})
	if err != nil {
		panic("couldn't connect to db: " + err.Error())
	}

	ins, err := pgxgen.Inspect(conn, "public")
	if err != nil {
		panic("error inspecting db: " + err.Error())
	}

	tmpl, err := template.New("a").Funcs(template.FuncMap{
		"exported": func(s ...string) string {
			var r string
			for k := range s {
				r += pgxgen.ExportedName(s[k])
			}
			return r
		},
	}).ParseGlob("./tmpl/*.tpl")
	if err != nil {
		panic("error reading templates: " + err.Error())
	}

	// Write enums
	for _, en := range ins.Enums {
		filename := filepath.Join(outdir, "pgxgen_enum_"+strings.ToLower(en.Name)+".go")
		f, err := os.Create(filename)
		if err != nil {
			f.Close()
			panic("error creating file: " + filename + ": " + err.Error())
		}
		err = tmpl.ExecuteTemplate(f, "enum.tpl",
			struct {
				PackageName string
				Enum        *pgxgen.Enum
			}{
				PackageName: pkgName,
				Enum:        en,
			})
		if err != nil {
			f.Close()
			panic("error executing template: " + filename + ": " + err.Error())
		}
		f.Close()
	}

	// Write tables
	for _, en := range ins.Tables {
		filename := filepath.Join(outdir, "pgxgen_table_"+strings.ToLower(en.Name)+".go")
		f, err := os.Create(filename)
		if err != nil {
			f.Close()
			panic("error creating file: " + filename + ": " + err.Error())
		}
		err = tmpl.ExecuteTemplate(f, "table.tpl",
			struct {
				PackageName string
				Table       *pgxgen.Table
			}{
				PackageName: pkgName,
				Table:       en,
			})
		if err != nil {
			f.Close()
			panic("error executing template: " + filename + ": " + err.Error())
		}
		f.Close()
	}

	// Write utils file which contains helpers
	filename := filepath.Join(outdir, "pgxgen_utils.go")
	f, err := os.Create(filename)
	if err != nil {
		f.Close()
		panic("error creating file: " + filename + ": " + err.Error())
	}
	err = tmpl.ExecuteTemplate(f, "utils.tpl",
		struct {
			PackageName string
		}{
			PackageName: pkgName,
		})
	if err != nil {
		f.Close()
		panic("error executing template: " + filename + ": " + err.Error())
	}
	f.Close()

	// Format output
	out, err := exec.Command("sh", "-c", "goimports -w "+filepath.Join(outdir, "*.go")).Output()
	if err != nil {
		panic("error formatting: " + err.Error() + "\n" + string(out))
	}
}
