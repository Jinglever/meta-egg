package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	bizgen "meta-egg/internal/domain/biz_generator"
	cmdgen "meta-egg/internal/domain/cmd_generator"
	commongen "meta-egg/internal/domain/common_generator"
	cfggen "meta-egg/internal/domain/config_generator"
	csttplgen "meta-egg/internal/domain/custom_template_generator"
	domaingen "meta-egg/internal/domain/domain_generator"
	handlergen "meta-egg/internal/domain/handler_generator"
	modelgen "meta-egg/internal/domain/model_generator"
	"meta-egg/internal/domain/modeler"
	pkggen "meta-egg/internal/domain/pkg_generator"
	projgen "meta-egg/internal/domain/project_generator"
	protogen "meta-egg/internal/domain/proto_generator"
	repogen "meta-egg/internal/domain/repo_generator"
	svcgen "meta-egg/internal/domain/server_generator"
	"meta-egg/internal/repo"
	repocfg "meta-egg/internal/repo/config"

	"github.com/gobwas/glob"
	log "github.com/sirupsen/logrus"

	jgcmd "github.com/Jinglever/go-command"
	jgfile "github.com/Jinglever/go-file"

	"github.com/urfave/cli/v2"
)

func updateProject(c *cli.Context) error {
	var err error
	checkDebugMode(c)
	cfg := loadEnvConfig(c)
	tryUncertain := c.Bool(FlagUncertain)
	customTemplateRoot := c.String(FlagTemplate)

	// parse xml file
	m, err := modeler.ParseXMLFile(cfg.Manifest.File)
	if err != nil {
		c.App.Writer.Write([]byte(fmt.Sprintf("parse xml file failed: %v\n", err)))
		return err
	}
	// ignore files
	ignoreFiles := make(map[string]bool)
	for _, file := range cfg.IgnoreFiles {
		ignoreFiles[file] = true
	}

	// tmp目录
	tmpRoot := filepath.Join(cfg.Manifest.Root, "generated")
	if err = os.RemoveAll(tmpRoot); err != nil {
		c.App.Writer.Write([]byte(fmt.Sprintf("fail to remove %s, err: %v\n", tmpRoot, err)))
		return err
	}
	if err = os.MkdirAll(tmpRoot, 0755); err != nil {
		c.App.Writer.Write([]byte(fmt.Sprintf("fail to create %s, err: %v\n", tmpRoot, err)))
		return err
	}

	// bak目录
	bakRoot := filepath.Join(cfg.Manifest.Root, "generated", "bak")
	if err = os.RemoveAll(bakRoot); err != nil {
		c.App.Writer.Write([]byte(fmt.Sprintf("fail to remove %s, err: %v\n", bakRoot, err)))
		return err
	}
	if err = os.MkdirAll(bakRoot, 0755); err != nil {
		c.App.Writer.Write([]byte(fmt.Sprintf("fail to create %s, err: %v\n", bakRoot, err)))
		return err
	}

	relativeDir2NeedConfirm := make(map[string]bool)
	var tmpRD2NC map[string]bool

	// project
	if err = projgen.Generate(tmpRoot, m.Project, projgen.ExtendParam{}); err != nil {
		c.App.Writer.Write([]byte(fmt.Sprintf("failed to generate project: %v\n", err)))
		return err
	}
	appedRelativeDir2NeedConfirm(relativeDir2NeedConfirm,
		map[string]bool{
			"":              true, // project root
			"_manifest":     false,
			"build/package": true,
		},
	) // special

	// update proto
	tmpRD2NC, err = protogen.Generate(tmpRoot, m.Project)
	if err != nil {
		c.App.Writer.Write([]byte(fmt.Sprintf("fail to generate proto, err: %v\n", err)))
		return err
	}
	appedRelativeDir2NeedConfirm(relativeDir2NeedConfirm, tmpRD2NC)

	// update pkg
	tmpRD2NC, err = pkggen.Generate(tmpRoot, m.Project)
	if err != nil {
		c.App.Writer.Write([]byte(fmt.Sprintf("fail to generate pkg, err: %v\n", err)))
		return err
	}
	appedRelativeDir2NeedConfirm(relativeDir2NeedConfirm, tmpRD2NC)

	// update common
	tmpRD2NC, err = commongen.Generate(tmpRoot, m.Project)
	if err != nil {
		c.App.Writer.Write([]byte(fmt.Sprintf("fail to generate common, err: %v\n", err)))
		return err
	}
	appedRelativeDir2NeedConfirm(relativeDir2NeedConfirm, tmpRD2NC)

	var dbOper repo.DBOperator
	if m.Project.Database != nil {
		dbName := m.Project.Database.Name
		if cfg.DB.DBName != "" {
			dbName = cfg.DB.DBName
		}
		dbOper, err = repo.NewDBOperator(m.Project.Database.Type,
			repocfg.DBConfig{
				Host:     cfg.DB.Host,
				Port:     cfg.DB.Port,
				User:     cfg.DB.User,
				Password: cfg.DB.Password,
				DBName:   dbName,
			}, cfg.IgnoreTables)
		if err != nil {
			c.App.Writer.Write([]byte(fmt.Sprintf("fail to create db operator, err: %v\n", err)))
			return err
		}
		defer dbOper.Close()
	}

	// update model
	tmpRD2NC, err = modelgen.Generate(tmpRoot, m.Project, dbOper)
	if err != nil {
		c.App.Writer.Write([]byte(fmt.Sprintf("fail to generate model, err: %v\n", err)))
		return err
	}
	appedRelativeDir2NeedConfirm(relativeDir2NeedConfirm, tmpRD2NC)

	// update repo
	tmpRD2NC, err = repogen.Generate(tmpRoot, m.Project, dbOper)
	if err != nil {
		c.App.Writer.Write([]byte(fmt.Sprintf("fail to generate model, err: %v\n", err)))
		return err
	}
	appedRelativeDir2NeedConfirm(relativeDir2NeedConfirm, tmpRD2NC)

	// biz
	tmpRD2NC, err = bizgen.Generate(tmpRoot, m.Project)
	if err != nil {
		c.App.Writer.Write([]byte(fmt.Sprintf("fail to generate biz, err: %v\n", err)))
		return err
	}
	appedRelativeDir2NeedConfirm(relativeDir2NeedConfirm, tmpRD2NC)

	// update domain
	tmpRD2NC, err = domaingen.Generate(tmpRoot, m.Project)
	if err != nil {
		c.App.Writer.Write([]byte(fmt.Sprintf("fail to generate domain, err: %v\n", err)))
		return err
	}
	appedRelativeDir2NeedConfirm(relativeDir2NeedConfirm, tmpRD2NC)

	// update handler
	tmpRD2NC, err = handlergen.Generate(tmpRoot, m.Project)
	if err != nil {
		c.App.Writer.Write([]byte(fmt.Sprintf("fail to generate handler, err: %v\n", err)))
		return err
	}
	appedRelativeDir2NeedConfirm(relativeDir2NeedConfirm, tmpRD2NC)

	// update server
	tmpRD2NC, err = svcgen.Generate(tmpRoot, m.Project)
	if err != nil {
		c.App.Writer.Write([]byte(fmt.Sprintf("fail to generate server, err: %v\n", err)))
		return err
	}
	appedRelativeDir2NeedConfirm(relativeDir2NeedConfirm, tmpRD2NC)

	// update config
	tmpRD2NC, err = cfggen.Generate(tmpRoot, m.Project)
	if err != nil {
		c.App.Writer.Write([]byte(fmt.Sprintf("fail to generate config, err: %v\n", err)))
		return err
	}
	appedRelativeDir2NeedConfirm(relativeDir2NeedConfirm, tmpRD2NC)

	// update cmd
	tmpRD2NC, err = cmdgen.Generate(tmpRoot, m.Project)
	if err != nil {
		c.App.Writer.Write([]byte(fmt.Sprintf("fail to generate cmd, err: %v\n", err)))
		return err
	}
	appedRelativeDir2NeedConfirm(relativeDir2NeedConfirm, tmpRD2NC)

	// apply custom template generator
	if customTemplateRoot != "" {
		tmpRD2NC, err = csttplgen.Generate(tmpRoot, m.Project, customTemplateRoot)
		if err != nil {
			c.App.Writer.Write([]byte(fmt.Sprintf("fail to apply custom template generator, err: %v\n", err)))
			return err
		}
		appedRelativeDir2NeedConfirm(relativeDir2NeedConfirm, tmpRD2NC)
	}

	// replace
	_, err = replaceCode(c, tmpRoot, cfg.Project.Root, bakRoot, relativeDir2NeedConfirm, tryUncertain, ignoreFiles)
	if err != nil {
		return err
	}
	return nil
}

func appedRelativeDir2NeedConfirm(a, b map[string]bool) {
	for k, v := range b {
		a[k] = v
	}
}

// 替换代码
// relativeDir2NeedConfirm: key是相对于工程根目录的路径，值是布尔值，代表是否需要确认；注意，不会递归查看子目录
// 需要确认才能替换的文件，替换时会先备份到bakRoot
// 只有当tryUncertain为true时，才会尝试替换需要确认的文件，否则只替换无需确认的文件
func replaceCode(c *cli.Context, srcRoot, targetRoot, bakRoot string,
	relativeDir2NeedConfirm map[string]bool, tryUncertain bool, ignoreFiles map[string]bool,
) (cnt int, err error) {
	ignoreFilePattern := make([]glob.Glob, 0, len(ignoreFiles))
	for k := range ignoreFiles {
		if strings.Contains(k, "*") {
			ignoreFilePattern = append(ignoreFilePattern, glob.MustCompile(k))
		}
	}
	chkIgnoreFilePattern := func(relativePath string) bool {
		for _, g := range ignoreFilePattern {
			if ok := g.Match(relativePath); ok {
				return true
			}
		}
		return false
	}
	noConfirmRelativeDir2Files := make(map[string][]string, 0)   // 无需确认就可以替换的文件
	needConfirmRelativeDir2Files := make(map[string][]string, 0) // 需要确认才能替换的文件
	for relativeDir, needConfirm := range relativeDir2NeedConfirm {
		// root
		srcDir := filepath.Join(srcRoot, relativeDir)
		if !jgfile.IsDir(srcDir) {
			continue
		}
		targetDir := filepath.Join(targetRoot, relativeDir)
		if err = os.MkdirAll(targetDir, 0755); err != nil {
			c.App.Writer.Write([]byte(fmt.Sprintf("fail to create target dir %s, err: %v\n", targetDir, err)))
			return cnt, err
		}
		// scan files src root
		files, err := os.ReadDir(srcDir)
		if err != nil {
			c.App.Writer.Write([]byte(fmt.Sprintf("fail to read src dir %s, err: %v\n", srcDir, err)))
			return cnt, err
		}
		// check if need replace
		for _, file := range files {
			if file.IsDir() {
				continue
			}
			if relativeDir == "_manifest" && file.Name() != "meta_egg.dtd" { // only check meta_egg.dtd
				continue
			}
			srcFile := filepath.Join(srcDir, file.Name())
			targetFile := filepath.Join(targetDir, file.Name())
			if jgfile.IsFile(targetFile) {
				// compare file content
				same, err := isFilesEqual(srcFile, targetFile)
				if err != nil {
					log.Errorf("fail to compare file %s vs %s, err: %v", srcFile, targetFile, err)
					return cnt, err
				}
				if same {
					continue
				}
				if needConfirm {
					if tryUncertain {
						if _, ok := needConfirmRelativeDir2Files[relativeDir]; !ok {
							needConfirmRelativeDir2Files[relativeDir] = make([]string, 0)
						}
						if !ignoreFiles[filepath.Join(relativeDir, file.Name())] &&
							!chkIgnoreFilePattern(filepath.Join(relativeDir, file.Name())) {
							needConfirmRelativeDir2Files[relativeDir] = append(needConfirmRelativeDir2Files[relativeDir], file.Name())
						}
					}
					continue
				}
			} else {
				if needConfirm {
					if tryUncertain {
						if _, ok := needConfirmRelativeDir2Files[relativeDir]; !ok {
							needConfirmRelativeDir2Files[relativeDir] = make([]string, 0)
						}
						// 对于新增的文件, 不受ignoreFiles里的通配规则的影响, 但仍受ignoreFiles里的具体文件名的影响
						if !ignoreFiles[filepath.Join(relativeDir, file.Name())] {
							needConfirmRelativeDir2Files[relativeDir] = append(needConfirmRelativeDir2Files[relativeDir], file.Name())
						}
					}
					continue
				}
			}
			if _, ok := noConfirmRelativeDir2Files[relativeDir]; !ok {
				noConfirmRelativeDir2Files[relativeDir] = make([]string, 0)
			}
			noConfirmRelativeDir2Files[relativeDir] = append(noConfirmRelativeDir2Files[relativeDir], file.Name())
		}
	}

	for relativeDir, files := range noConfirmRelativeDir2Files {
		srcDir := filepath.Join(srcRoot, relativeDir)
		targetDir := filepath.Join(targetRoot, relativeDir)
		c.App.Writer.Write([]byte(fmt.Sprintf("replaced files in %s\n", relativeDir)))

		for _, file := range files {
			srcFile := filepath.Join(srcDir, file)
			targetFile := filepath.Join(targetDir, file)

			// 无需确认的，且不在ignoreFiles中的文件，才会被替换
			if ignoreFiles[filepath.Join(relativeDir, file)] {
				continue
			}

			// copy
			if _, err := jgfile.CopyFile(srcFile, targetFile); err != nil {
				c.App.Writer.Write([]byte(fmt.Sprintf("fail to copy file %s to %s, err: %v\n", srcFile, targetFile, err)))
				return cnt, err
			}
			c.App.Writer.Write([]byte(fmt.Sprintf("\t%s%s%s [%s]\n", ColorFileDone, file, ColorEnd, GreenCheck)))
			cnt++
		}
	}

	if len(needConfirmRelativeDir2Files) == 0 {
		c.App.Writer.Write([]byte(fmt.Sprintf("replace %v files\n", cnt)))
		return cnt, nil
	}

	for relativeDir, files := range needConfirmRelativeDir2Files {
		if len(files) == 0 {
			continue
		}

		var (
			replaceMap = make([]string, 0)
			newMap     = make([]string, 0)
			baseGo     string
		)

		bakDir := filepath.Join(bakRoot, relativeDir)
		if err = os.MkdirAll(bakDir, 0755); err != nil {
			c.App.Writer.Write([]byte(fmt.Sprintf("fail to create bak dir %s, err: %v\n", bakDir, err)))
			return cnt, err
		}
		srcDir := filepath.Join(srcRoot, relativeDir)
		targetDir := filepath.Join(targetRoot, relativeDir)
		for _, file := range files {
			if file == "base.go" {
				baseGo = file
				continue
			}

			targetFile := filepath.Join(targetDir, file)

			if jgfile.IsFile(targetFile) {
				replaceMap = append(replaceMap, file)
			} else {
				newMap = append(newMap, file)
			}
		}

		if len(replaceMap) > 0 {
			c.App.Writer.Write([]byte(fmt.Sprintf("%s%sFound %sdifferent%s %s%sfiles in%s %s%s%s\n",
				ColorStatementDiff, FontItalic, FontBold, ColorEnd,
				ColorStatementDiff, FontItalic, ColorEnd,
				ColorRelativeDir, relativeDir, ColorEnd,
			)))
			for idx, file := range replaceMap {
				if (idx+1)%2 == 0 || idx == len(replaceMap)-1 {
					c.App.Writer.Write([]byte(fmt.Sprintf("\t%s%s%s\n", ColorFilesDiff, file, ColorEnd)))
				} else {
					c.App.Writer.Write([]byte(fmt.Sprintf("\t%s%s%s", ColorFilesDiff, file, ColorEnd)))
				}
			}

			// confirm replace
			opt := jgcmd.AskForOption(
				ColorStatementDiff+FontItalic+"confirm to replace into project?"+ColorEnd,
				[]string{"y", "n"},
				"",
			)
			if opt == "y" {
				for _, file := range replaceMap {
					srcFile := filepath.Join(srcDir, file)
					targetFile := filepath.Join(targetDir, file)

					// backup
					bakFile := filepath.Join(bakDir, file)
					if err = os.Rename(targetFile, bakFile); err != nil {
						c.App.Writer.Write([]byte(fmt.Sprintf("fail to backup file %s to %s, err: %v\n", targetFile, bakFile, err)))
						return cnt, err
					}
					// copy
					if _, err := jgfile.CopyFile(srcFile, targetFile); err != nil {
						c.App.Writer.Write([]byte(fmt.Sprintf("fail to copy file %s to %s, err: %v\n", srcFile, targetFile, err)))
						return cnt, err
					}
					c.App.Writer.Write([]byte(fmt.Sprintf("\t%s%s%s [%s]\n", ColorFileDone, file, ColorEnd, GreenCheck)))
					cnt++
				}
			}
		}
		if len(newMap) > 0 {
			c.App.Writer.Write([]byte(fmt.Sprintf("%s%sFound %snew%s %s%sfiles in%s %s%s%s\n",
				ColorStatementNew, FontItalic, FontBold, ColorEnd,
				ColorStatementNew, FontItalic, ColorEnd,
				ColorRelativeDir, relativeDir, ColorEnd,
			)))
			for idx, file := range newMap {
				if (idx+1)%2 == 0 || idx == len(newMap)-1 {
					c.App.Writer.Write([]byte(fmt.Sprintf("\t%s%s%s\n", ColorFilesNew, file, ColorEnd)))
				} else {
					c.App.Writer.Write([]byte(fmt.Sprintf("\t%s%s%s", ColorFilesNew, file, ColorEnd)))
				}
			}

			// confirm replace
			opt := jgcmd.AskForOption(
				ColorStatementNew+FontItalic+"confirm to copy into project?"+ColorEnd,
				[]string{"y", "n"},
				"",
			)
			if opt == "y" {
				for _, file := range newMap {
					srcFile := filepath.Join(srcDir, file)
					targetFile := filepath.Join(targetDir, file)

					// copy
					if _, err := jgfile.CopyFile(srcFile, targetFile); err != nil {
						c.App.Writer.Write([]byte(fmt.Sprintf("fail to copy file %s to %s, err: %v\n", srcFile, targetFile, err)))
						return cnt, err
					}
					c.App.Writer.Write([]byte(fmt.Sprintf("\t%s%s%s [%s]\n", ColorFileDone, file, ColorEnd, GreenCheck)))
					cnt++
				}
			}
		}
		if baseGo != "" {
			c.App.Writer.Write([]byte(fmt.Sprintf("%s%sFound %sdifferent%s %sbase.go%s %s%sin%s %s%s%s\n",
				ColorStatementBase, FontItalic, FontBold, ColorEnd,
				ColorFilesBase, ColorEnd,
				ColorStatementBase, FontItalic, ColorEnd,
				ColorRelativeDir, relativeDir, ColorEnd,
			)))
			// confirm replace
			opt := jgcmd.AskForOption(
				ColorStatementBase+FontItalic+"confirm to replace into project?"+ColorEnd,
				[]string{"y", "n"},
				"",
			)
			if opt == "y" {
				file := baseGo
				srcFile := filepath.Join(srcDir, file)
				targetFile := filepath.Join(targetDir, file)

				// backup
				if jgfile.IsFile(targetFile) {
					bakFile := filepath.Join(bakDir, file)
					if err = os.Rename(targetFile, bakFile); err != nil {
						c.App.Writer.Write([]byte(fmt.Sprintf("fail to backup file %s to %s, err: %v\n", targetFile, bakFile, err)))
						return cnt, err
					}
				}
				// copy
				if _, err := jgfile.CopyFile(srcFile, targetFile); err != nil {
					c.App.Writer.Write([]byte(fmt.Sprintf("fail to copy file %s to %s, err: %v\n", srcFile, targetFile, err)))
					return cnt, err
				}
				c.App.Writer.Write([]byte(fmt.Sprintf("\t%s%s%s [%s]\n", ColorFileDone, file, ColorEnd, GreenCheck)))
				cnt++
			}
		}
	}
	c.App.Writer.Write([]byte(fmt.Sprintf("replace/new %v files\n", cnt)))
	return cnt, nil
}

// 排除抬头注释的影响，比对两个文件的内容是否相同
func isFilesEqual(path1, path2 string) (bool, error) {
	b1, err := os.ReadFile(path1)
	if err != nil {
		log.Errorf("read file %s failed: %v", path1, err)
		return false, err
	}
	b2, err := os.ReadFile(path2)
	if err != nil {
		log.Errorf("read file %s failed: %v", path2, err)
		return false, err
	}
	// Define a regular expression to match the lines to remove.
	removeRegex := regexp.MustCompile(`(?m)^\s*(//\s*)?(Version|Generated\s*at):\s*.*$`)

	// Remove the matching lines from the text.
	cleanedStr1 := removeRegex.ReplaceAllString(string(b1), "")
	cleanedStr2 := removeRegex.ReplaceAllString(string(b2), "")
	return cleanedStr1 == cleanedStr2, nil
}
