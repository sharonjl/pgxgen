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
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"github.com/jackc/pgx"
	"github.com/mitchellh/go-homedir"
	"github.com/pelletier/go-toml"
	"github.com/sharonjl/pgxgen"
	"github.com/sharonjl/pgxgen/tmpl"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tangzero/inflector"
)

var cfgFile string
var dbHost string
var dbUser string
var dbPassword string
var dbName string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "pgxgen",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		outf := cmd.Flag("out").Value.String()
		gendir, err := filepath.Abs(outf)
		if err != nil {
			panic("output directory: " + outf + ": " + err.Error())
		}

		modelRunFn(gendir, cmd, args)
		//qbRunFn(gendir, cmd, args)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.pgxgen.yaml)")
	rootCmd.PersistentFlags().StringVar(&dbHost, "dbHost", "localhost", "connection hostname")
	rootCmd.PersistentFlags().StringVar(&dbUser, "dbUser", "sharon", "connecting user")
	rootCmd.PersistentFlags().StringVar(&dbPassword, "dbPassword", "", "password for connecting user")
	rootCmd.PersistentFlags().StringVar(&dbName, "dbName", "", "database to connect to")
	rootCmd.PersistentFlags().String("package", "dbmodel", "package name")
	rootCmd.PersistentFlags().String("query", "config.toml", "query definition file")
	rootCmd.PersistentFlags().String("out", ".", "output")
	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".pgxgen" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".pgxgen")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func qbRunFn(gendir string, cmd *cobra.Command, args []string) {
	// Output directory
	outf := filepath.Join(gendir, "builder")
	outdir, err := filepath.Abs(outf)
	if err != nil {
		panic("output directory: " + outf + ": " + err.Error())
	}

	err = os.MkdirAll(outf, os.ModePerm)
	if err != nil {
		panic("error creating output directory: " + outf + ": " + err.Error())
	}

	// Package name
	pkgName := "builder"

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
		"inc": func(i int) string {
			return strconv.Itoa(i + 1)
		},
	}).ParseGlob("./tmpl/*.tpl")
	if err != nil {
		panic("error reading templates: " + err.Error())
	}

	// Write enums
	// for _, en := range ins.Enums {
	// 	filename := filepath.Join(outdir, "pgxgen_enum_"+strings.ToLower(en.Name)+".go")
	// 	f, err := os.Create(filename)
	// 	if err != nil {
	// 		f.Close()
	// 		panic("error creating file: " + filename + ": " + err.Error())
	// 	}
	// 	err = tmpl.ExecuteTemplate(f, "enum.tpl",
	// 		struct {
	// 			PackageName string
	// 			Enum        *pgxgen.Enum
	// 		}{
	// 			PackageName: pkgName,
	// 			Enum:        en,
	// 		})
	// 	if err != nil {
	// 		f.Close()
	// 		panic("error executing template: " + filename + ": " + err.Error())
	// 	}
	// 	f.Close()
	// }

	// Write tables
	for _, en := range ins.Tables {
		filename := filepath.Join(outdir, "qb_table_"+strings.ToLower(en.Name)+".go")
		f, err := os.Create(filename)
		if err != nil {
			f.Close()
			panic("error creating file: " + filename + ": " + err.Error())
		}
		err = tmpl.ExecuteTemplate(f, "table_qb.tpl",
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
	filename := filepath.Join(outdir, "qb_utils.go")
	f, err := os.Create(filename)
	if err != nil {
		f.Close()
		panic("error creating file: " + filename + ": " + err.Error())
	}
	var tables []*pgxgen.Table
	for _, t := range ins.Tables {
		tables = append(tables, t)
	}
	err = tmpl.ExecuteTemplate(f, "utils_qb.tpl",
		struct {
			PackageName string
			Tables      []*pgxgen.Table
		}{
			PackageName: pkgName,
			Tables:      tables,
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

func modelRunFn(gendir string, cmd *cobra.Command, args []string) {

	// Output directory
	// Output directory
	outf := filepath.Join(gendir)
	outdir, err := filepath.Abs(outf)
	if err != nil {
		panic("output directory: " + outf + ": " + err.Error())
	}

	err = os.MkdirAll(outf, os.ModePerm)
	if err != nil {
		panic("error creating output directory: " + outf + ": " + err.Error())
	}

	modelDir := filepath.Join(outdir, cmd.Flag("package").Value.String())
	err = os.MkdirAll(modelDir, os.ModePerm)
	if err != nil {
		panic("error creating models directory: " + modelDir + ": " + err.Error())
	}

	storedir := filepath.Join(outdir, "store")
	postgresImplDir := filepath.Join(storedir, "postgres")
	err = os.MkdirAll(postgresImplDir, os.ModePerm)
	if err != nil {
		panic("error creating output fn directory: " + postgresImplDir + ": " + err.Error())
	}

	srcPath := []rune(filepath.Join(os.Getenv("GOPATH"), "src"))
	importPath := string([]rune(outdir)[len(srcPath)+1:])

	modelPkgName := filepath.Base(modelDir)

	// Read QueryDefinition defn
	queryFile := cmd.Flag("query").Value.String()
	queryFile, _ = filepath.Abs(queryFile)

	queryDefs, err := toml.LoadFile(queryFile)
	if err != nil {
		panic("error could not read query defns: " + queryFile + ": " + err.Error())
	}

	// Read DB
	conn, err := pgx.Connect(pgx.ConnConfig{
		Host:     dbHost,
		User:     dbUser,
		Password: dbPassword,
		Database: dbName,
	})
	if err != nil {
		panic("couldn't connect to db: " + err.Error())
	}

	pgdata, err := pgxgen.Inspect(conn, "public")
	if err != nil {
		panic("error inspecting db: " + err.Error())
	}

	// Read query config
	queryDoc := pgxgen.QueryDefinitions{}
	err = queryDefs.Unmarshal(&queryDoc)
	if err != nil {
		panic("error unmarshalling queries: " + err.Error())
	}
	queries := pgxgen.ProcessQueryDefinitions(queryDoc, *pgdata)

	tpl := template.New("model").Funcs(template.FuncMap{
		"exported": func(s ...string) string {
			var r string
			for k := range s {
				r += pgxgen.ExportedName(s[k])
			}
			return r
		},
		"pluralize": func(s string) string {
			return inflector.Pluralize(s)
		},
		"inc": func(i int) string {
			return strconv.Itoa(i + 1)
		},
	})

	tpl, _ = tpl.New("enum.tpl").Parse(string(tmpl.MustAsset("../tmpl/enum.tpl")))
	tpl, _ = tpl.New("table.tpl").Parse(string(tmpl.MustAsset("../tmpl/table.tpl")))
	tpl, _ = tpl.New("table_fn.tpl").Parse(string(tmpl.MustAsset("../tmpl/table_fn.tpl")))
	tpl, _ = tpl.New("utils.tpl").Parse(string(tmpl.MustAsset("../tmpl/utils.tpl")))
	tpl, _ = tpl.New("store.tpl").Parse(string(tmpl.MustAsset("../tmpl/store.tpl")))
	tpl, _ = tpl.New("queries.tpl").Parse(string(tmpl.MustAsset("../tmpl/queries.tpl")))
	tpl, _ = tpl.New("postgres.tpl").Parse(string(tmpl.MustAsset("../tmpl/postgres.tpl")))
	tpl, _ = tpl.New("store_keys.tpl").Parse(string(tmpl.MustAsset("../tmpl/store_keys.tpl")))

	// Write enums
	for _, en := range pgdata.Enums {
		filename := filepath.Join(modelDir, strings.ToLower(en.Name)+".pgxgen.go")
		f, err := os.Create(filename)
		if err != nil {
			f.Close()
			panic("error creating file: " + filename + ": " + err.Error())
		}
		err = tpl.ExecuteTemplate(f, "enum.tpl",
			struct {
				PackageName string
				ImportPath  string
				Enum        *pgxgen.Enum
			}{
				PackageName: modelPkgName,
				ImportPath:  importPath,
				Enum:        en,
			})
		if err != nil {
			f.Close()
			panic("error executing template: " + filename + ": " + err.Error())
		}
		f.Close()
	}

	// Write tables
	for _, en := range pgdata.Tables {
		// Model
		filename := filepath.Join(modelDir, strings.ToLower(en.Name)+".pgxgen.go")
		f, err := os.Create(filename)
		if err != nil {
			f.Close()
			panic("error creating file: " + filename + ": " + err.Error())
		}
		err = tpl.ExecuteTemplate(f, "table.tpl",
			struct {
				PackageName string
				ImportPath  string
				Table       *pgxgen.Table
			}{
				PackageName: modelPkgName,
				ImportPath:  importPath,
				Table:       en,
			})
		if err != nil {
			f.Close()
			panic("error executing template: " + filename + ": " + err.Error())
		}
		f.Close()

		filename = filepath.Join(postgresImplDir, strings.ToLower(en.Name)+".pgxgen.go")
		f, err = os.Create(filename)
		if err != nil {
			f.Close()
			panic("error creating file: " + filename + ": " + err.Error())
		}
		err = tpl.ExecuteTemplate(f, "table_fn.tpl",
			struct {
				PackageName      string
				ImportPath       string
				ModelPackageName string
				Table            *pgxgen.Table
			}{
				ModelPackageName: modelPkgName,
				ImportPath:       importPath,
				PackageName:      "postgres",
				Table:            en,
			})
		if err != nil {
			f.Close()
			panic("error executing template: " + filename + ": " + err.Error())
		}
		f.Close()
	}

	// Write queries
	{
		filename := filepath.Join(postgresImplDir, "queries.pgxgen.go")
		f, err := os.Create(filename)
		if err != nil {
			f.Close()
			panic("error creating file: " + filename + ": " + err.Error())
		}
		err = tpl.ExecuteTemplate(f, "queries.tpl",
			struct {
				PackageName      string
				ImportPath       string
				ModelPackageName string
				Queries          []pgxgen.Query
			}{
				PackageName:      "postgres",
				ModelPackageName: modelPkgName,
				ImportPath:       importPath,
				Queries:          queries,
			})
		if err != nil {
			f.Close()
			panic("error executing template: " + filename + ": " + err.Error())
		}
		f.Close()
	}
	{
		filename := filepath.Join(postgresImplDir, "postgres.pgxgen.go")
		f, err := os.Create(filename)
		if err != nil {
			f.Close()
			panic("error creating file: " + filename + ": " + err.Error())
		}
		err = tpl.ExecuteTemplate(f, "postgres.tpl",
			struct {
				PackageName      string
				ImportPath       string
				ModelPackageName string
				Queries          []pgxgen.Query
			}{
				PackageName:      "postgres",
				ModelPackageName: modelPkgName,
				ImportPath:       importPath,
				Queries:          queries,
			})
		if err != nil {
			f.Close()
			panic("error executing template: " + filename + ": " + err.Error())
		}
		f.Close()
	}

	{
		filename := filepath.Join(storedir, "store.pgxgen.go")
		f, err := os.Create(filename)
		if err != nil {
			f.Close()
			panic("error creating file: " + filename + ": " + err.Error())
		}
		err = tpl.ExecuteTemplate(f, "store.tpl",
			struct {
				PackageName string
				ImportPath  string
			}{
				PackageName: "store",
				ImportPath:  importPath,
			})
		if err != nil {
			f.Close()
			panic("error executing template: " + filename + ": " + err.Error())
		}
		f.Close()

		filename = filepath.Join(storedir, "keys.pgxgen.go")
		f, err = os.Create(filename)
		if err != nil {
			f.Close()
			panic("error creating file: " + filename + ": " + err.Error())
		}
		err = tpl.ExecuteTemplate(f, "store_keys.tpl",
			struct {
				PackageName      string
				ImportPath       string
				ModelPackageName string
				Queries          []pgxgen.Query
			}{
				PackageName:      "store",
				ModelPackageName: modelPkgName,
				ImportPath:       importPath,
				Queries:          queries,
			})
		if err != nil {
			f.Close()
			panic("error executing template: " + filename + ": " + err.Error())
		}
		f.Close()
	}

	// Write utils file which contains helpers
	//filename := filepath.Join(modelDir, "pgxgen_utils.go")
	//f, err := os.Create(filename)
	//if err != nil {
	//	f.Close()
	//	panic("error creating file: " + filename + ": " + err.Error())
	//}
	//err = tpl.ExecuteTemplate(f, "utils.tpl",
	//	struct {
	//		PackageName string
	//	}{
	//		PackageName: modelPkgName,
	//	})
	//if err != nil {
	//	f.Close()
	//	panic("error executing template: " + filename + ": " + err.Error())
	//}
	//f.Close()

	// Format output
	out, err := exec.Command("sh", "-c", "goimports -w "+filepath.Join(modelDir, "./..")).Output()
	if err != nil {
		panic("error formatting: " + err.Error() + "\n" + string(out))
	}
}
