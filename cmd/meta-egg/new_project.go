package main

import (
	"fmt"
	"os"
	"path/filepath"

	"meta-egg/internal/config"
	bizgen "meta-egg/internal/domain/biz_generator"
	cmdgen "meta-egg/internal/domain/cmd_generator"
	commongen "meta-egg/internal/domain/common_generator"
	cfggen "meta-egg/internal/domain/config_generator"
	csttplgen "meta-egg/internal/domain/custom_template_generator"
	domaingen "meta-egg/internal/domain/domain_generator"
	handlergen "meta-egg/internal/domain/handler_generator"
	"meta-egg/internal/domain/helper"
	modelgen "meta-egg/internal/domain/model_generator"
	"meta-egg/internal/domain/modeler"
	pkggen "meta-egg/internal/domain/pkg_generator"
	projgen "meta-egg/internal/domain/project_generator"
	protogen "meta-egg/internal/domain/proto_generator"
	repogen "meta-egg/internal/domain/repo_generator"
	svcgen "meta-egg/internal/domain/server_generator"
	"meta-egg/internal/model"

	jgcmd "github.com/Jinglever/go-command"
	"github.com/spf13/cobra"
)

func newProject(cmd *cobra.Command) error {
	checkDebugMode(cmd)
	customTemplateRoot, _ := cmd.Flags().GetString(FlagTemplate)

	proj := &model.Project{}

	// ask for project name
	proj.Name = jgcmd.AskForInput(
		"Please input project EN name",
		"",
	)
	proj.Name = helper.NormalizeProjectName(proj.Name)
	// ask for project description
	proj.Desc = jgcmd.AskForInput(
		"Please input project description",
		"",
	)
	// ask for go module name
	proj.GoModule = jgcmd.AskForInput(
		"Please input go module name",
		"",
	)
	// ask for go version
	proj.GoVersion = jgcmd.AskForInput(
		"Please input go version",
		"1.23.0",
	)
	// ask for server type
	proj.ServerType = model.ServerType(jgcmd.AskForOption(
		"Please select server type",
		[]string{"GRPC", "HTTP", "ALL"},
		"ALL",
	))

	var opt string
	ep := projgen.ExtendParam{}

	// ask for if not need auth
	opt = jgcmd.AskForOption(
		"Do you need auth such as access token?",
		[]string{"y", "n"},
		"y",
	)
	if opt == "n" {
		proj.NoAuth = true
	}

	// ask for if need database
	opt = jgcmd.AskForOption(
		"Do you need database?",
		[]string{"y", "n"},
		"y",
	)
	if opt == "y" {
		ep.NeedDatabase = true

		// ask for which db type to use
		ep.DatabaseType = model.DatabaseType(jgcmd.AskForOption(
			"Please select database type",
			[]string{"MySQL", "PostgreSQL"},
			"PostgreSQL",
		))

		// ask for if need table demo
		opt = jgcmd.AskForOption(
			"Do you need table demo?",
			[]string{"y", "n"},
			"y",
		)
		if opt == "y" {
			ep.NeedTableDemo = true
		}
	}

	// get current directory
	codeDir, err := os.Getwd()
	if err != nil {
		cmd.OutOrStdout().Write([]byte(fmt.Sprintf("failed to get current directory: %v\n", err)))
		return err
	}
	projRoot := filepath.Join(codeDir, proj.Name)

	if err = projgen.Generate(projRoot, proj, ep); err != nil {
		cmd.OutOrStdout().Write([]byte(fmt.Sprintf("failed to generate project: %v\n", err)))
		return err
	}

	// continue
	envFile := filepath.Join(projRoot, "_manifest", "env.yml")
	cfg := config.LoadEnvFile(envFile)

	// parse xml file
	m, err := modeler.ParseXMLFile(cfg.Manifest.File)
	if err != nil {
		cmd.OutOrStdout().Write([]byte(fmt.Sprintf("parse xml file failed: %v\n", err)))
		return err
	}

	var rD2NC map[string]bool

	// proto
	rD2NC, err = protogen.Generate(projRoot, m.Project)
	if err != nil {
		cmd.OutOrStdout().Write([]byte(fmt.Sprintf("fail to generate proto, err: %v\n", err)))
		return err
	}
	printRD2NC(rD2NC, cmd)

	// pkg
	rD2NC, err = pkggen.Generate(projRoot, m.Project)
	if err != nil {
		cmd.OutOrStdout().Write([]byte(fmt.Sprintf("fail to generate pkg, err: %v\n", err)))
		return err
	}
	printRD2NC(rD2NC, cmd)

	// common
	rD2NC, err = commongen.Generate(projRoot, m.Project)
	if err != nil {
		cmd.OutOrStdout().Write([]byte(fmt.Sprintf("fail to generate common, err: %v\n", err)))
		return err
	}
	printRD2NC(rD2NC, cmd)

	// model
	rD2NC, err = modelgen.Generate(projRoot, m.Project, nil)
	if err != nil {
		cmd.OutOrStdout().Write([]byte(fmt.Sprintf("fail to generate model, err: %v\n", err)))
		return err
	}
	printRD2NC(rD2NC, cmd)

	// repo
	rD2NC, err = repogen.Generate(projRoot, m.Project, nil)
	if err != nil {
		cmd.OutOrStdout().Write([]byte(fmt.Sprintf("fail to generate model, err: %v\n", err)))
		return err
	}
	printRD2NC(rD2NC, cmd)

	// domain
	rD2NC, err = domaingen.Generate(projRoot, m.Project)
	if err != nil {
		cmd.OutOrStdout().Write([]byte(fmt.Sprintf("fail to generate domain, err: %v\n", err)))
		return err
	}
	printRD2NC(rD2NC, cmd)

	// biz
	rD2NC, err = bizgen.Generate(projRoot, m.Project)
	if err != nil {
		cmd.OutOrStdout().Write([]byte(fmt.Sprintf("fail to generate biz, err: %v\n", err)))
		return err
	}
	printRD2NC(rD2NC, cmd)

	// handler
	rD2NC, err = handlergen.Generate(projRoot, m.Project)
	if err != nil {
		cmd.OutOrStdout().Write([]byte(fmt.Sprintf("fail to generate handler, err: %v\n", err)))
		return err
	}
	printRD2NC(rD2NC, cmd)

	// server
	rD2NC, err = svcgen.Generate(projRoot, m.Project)
	if err != nil {
		cmd.OutOrStdout().Write([]byte(fmt.Sprintf("fail to generate server, err: %v\n", err)))
		return err
	}
	printRD2NC(rD2NC, cmd)

	// config
	rD2NC, err = cfggen.Generate(projRoot, m.Project)
	if err != nil {
		cmd.OutOrStdout().Write([]byte(fmt.Sprintf("fail to generate config, err: %v\n", err)))
		return err
	}
	printRD2NC(rD2NC, cmd)

	// cmd
	rD2NC, err = cmdgen.Generate(projRoot, m.Project)
	if err != nil {
		cmd.OutOrStdout().Write([]byte(fmt.Sprintf("fail to generate cmd, err: %v\n", err)))
		return err
	}
	printRD2NC(rD2NC, cmd)

	// apply custom template generator
	if customTemplateRoot != "" {
		_, err = csttplgen.Generate(projRoot, m.Project, customTemplateRoot)
		if err != nil {
			cmd.OutOrStdout().Write([]byte(fmt.Sprintf("fail to apply custom template generator, err: %v\n", err)))
			return err
		}
	}

	// print project generated successfully in green color
	cmd.OutOrStdout().Write([]byte("\033[32mProject generated successfully\033[0m\n"))
	return nil
}

func printRD2NC(rD2NC map[string]bool, cmd *cobra.Command) {
	for rd := range rD2NC {
		cmd.OutOrStdout().Write([]byte(fmt.Sprintf("  %s [\033[32m\u2713\033[0m]\n", rd)))
	}
}
