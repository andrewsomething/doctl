/*
Copyright 2018 The Doctl Authors All rights reserved.
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

package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/digitalocean/doctl/do"
	"gopkg.in/yaml.v3"
)

// ServerlessExtras adds commands to the 'serverless' subtree for which the cobra wrappers were autogenerated from
// oclif equivalents and subsequently modified.
func ServerlessExtras(cmd *Command) {

	create := CmdBuilder(cmd, RunServerlessExtraCreate, "init <path>", "Initialize a 'functions project' directory in your local file system",
		`The `+"`"+`doctl serverless init`+"`"+` command specifies a directory in your file system which will hold functions and
supporting artifacts while you're developing them.  This 'functions project' can be uploaded to your functions namespace for testing.
Later, after the functions project is committed to a `+"`"+`git`+"`"+` repository, you can create an app, or an app component, from it.

Type `+"`"+`doctl serverless status --languages`+"`"+` for a list of supported languages.  Use one of the displayed keywords
to choose your sample language for `+"`"+`doctl serverless init`+"`"+`.`,
		Writer)
	AddStringFlag(create, "language", "l", "javascript", "Language for the initial sample code")
	AddBoolFlag(create, "overwrite", "", false, "Clears and reuses an existing directory")

	deploy := CmdBuilder(cmd, RunServerlessExtraDeploy, "deploy <directory>", "Deploy a functions project to your functions namespace",
		`At any time you can use `+"`"+`doctl serverless deploy`+"`"+` to upload the contents of a functions project in your file system for
testing in your serverless namespace.  The project must be organized in the fashion expected by an App Platform Functions
component.  The `+"`"+`doctl serverless init`+"`"+` command will create a properly organized directory for you to work in.`,
		Writer)
	AddStringFlag(deploy, "env", "", "", "Path to runtime environment file")
	AddStringFlag(deploy, "build-env", "", "", "Path to build-time environment file")
	AddStringFlag(deploy, "apihost", "", "", "API host to use")
	AddStringFlag(deploy, "auth", "", "", "OpenWhisk auth token to use")
	AddBoolFlag(deploy, "insecure", "", false, "Ignore SSL Certificates")
	AddBoolFlag(deploy, "verbose-build", "", false, "Display build details")
	AddBoolFlag(deploy, "verbose-zip", "", false, "Display start/end of zipping phase for each function")
	AddBoolFlag(deploy, "yarn", "", false, "Use yarn instead of npm for node builds")
	AddStringFlag(deploy, "include", "", "", "Functions and/or packages to include")
	AddStringFlag(deploy, "exclude", "", "", "Functions and/or packages to exclude")
	AddBoolFlag(deploy, "remote-build", "", false, "Run builds remotely")
	AddBoolFlag(deploy, "incremental", "", false, "Deploy only changes since last deploy")

	getMetadata := CmdBuilder(cmd, RunServerlessExtraGetMetadata, "get-metadata <directory>", "Obtain metadata of a functions project",
		`The `+"`"+`doctl serverless get-metadata`+"`"+` command produces a JSON structure that summarizes the contents of a functions
project (a directory you have designated for functions development).  This can be useful for feeding into other tools.`,
		Writer)
	AddStringFlag(getMetadata, "env", "", "", "Path to environment file")
	AddStringFlag(getMetadata, "include", "", "", "Functions or packages to include")
	AddStringFlag(getMetadata, "exclude", "", "", "Functions or packages to exclude")
	AddBoolFlag(getMetadata, "project-reader", "", false, "Test new project reader service")
	getMetadata.Flags().MarkHidden("project-reader")

	watch := CmdBuilder(cmd, RunServerlessExtraWatch, "watch <directory>", "Watch a functions project directory, deploying incrementally on change",
		`Type `+"`"+`doctl serverless watch <directory>`+"`"+` in a separate terminal window.  It will run until interrupted.
It will watch the directory (which should be one you initialized for serverless development) and will deploy
the contents to the cloud incrementally as it detects changes.`,
		Writer)
	AddStringFlag(watch, "env", "", "", "Path to runtime environment file")
	AddStringFlag(watch, "build-env", "", "", "Path to build-time environment file")
	AddStringFlag(watch, "apihost", "", "", "API host to use")
	AddStringFlag(watch, "auth", "", "", "OpenWhisk auth token to use")
	AddBoolFlag(watch, "insecure", "", false, "Ignore SSL Certificates")
	AddBoolFlag(watch, "verbose-build", "", false, "Display build details")
	AddBoolFlag(watch, "verbose-zip", "", false, "Display start/end of zipping phase for each function")
	AddBoolFlag(watch, "yarn", "", false, "Use yarn instead of npm for node builds")
	AddStringFlag(watch, "include", "", "", "Functions and/or packages to include")
	AddStringFlag(watch, "exclude", "", "", "Functions and/or packages to exclude")
	AddBoolFlag(watch, "remote-build", "", false, "Run builds remotely")
}

// RunServerlessExtraCreate supports the 'serverless init' command
func RunServerlessExtraCreate(c *CmdConfig) error {
	if err := ensureOneArg(c); err != nil {
		return err
	}
	project := c.Args[0]
	overwrite, _ := c.Doit.GetBool(c.NS, "overwrite")
	language, _ := c.Doit.GetString(c.NS, "language")

	// Determine the kind and sample
	kind, sample, ts, err := languageToKindAndSample(c, language)
	if err != nil {
		return err
	}

	// Make the config and various paths
	config := configTemplate()
	configFile := filepath.Join(project, "project.yml")
	gitignoreFile := filepath.Join(project, ".gitignore")
	samplePackage := filepath.Join(project, "packages", "sample")

	// Prepare the project area
	if err = prepareProjectArea(project, overwrite); err != nil {
		return err
	}
	if err = doMkdir(samplePackage, true); err != nil {
		return err
	}

	// Generate the sample
	actionDir, err := generateSample(kind, &config, sample, samplePackage, ts)
	if err != nil {
		return err
	}

	// Write the config
	data, err := yaml.Marshal(&config)
	if err != nil {
		return err
	}
	if err = writeAFile(configFile, data); err != nil {
		return err
	}

	// Add the .gitignore
	ignores := gitignores
	if ts {
		ignores += ignoreForTypescript
	}
	if err = writeAFile(gitignoreFile, []byte(ignores)); err != nil {
		return err
	}

	// Add typescript-specific information
	if ts {
		pjFile := filepath.Join(actionDir, "package.json")
		if err = writeAFile(pjFile, []byte(packageJSONForTypescript)); err != nil {
			return err
		}
		tscFile := filepath.Join(actionDir, "tsconfig.json")
		if err = writeAFile(tscFile, []byte(tsconfigJSON)); err != nil {
			return err
		}
		includeFile := filepath.Join(actionDir, ".include")
		if err = writeAFile(includeFile, []byte("lib\n")); err != nil {
			return err
		}
	}

	// Print informational success message
	fmt.Fprintf(c.Out, `A local functions project directory '%s' was created for you.
You may deploy it by running the command shown on the next line:
  doctl serverless deploy %s
`, project, project)
	return nil
}

// RunServerlessExtraDeploy supports the 'serverless deploy' command
func RunServerlessExtraDeploy(c *CmdConfig) error {
	adjustIncludeAndExclude(c)
	err := ensureOneArg(c)
	if err != nil {
		return err
	}
	output, err := RunServerlessExec("deploy", c, []string{flagInsecure, flagVerboseBuild, flagVerboseZip, flagYarn, flagRemoteBuild, flagIncremental},
		[]string{flagEnv, flagBuildEnv, flagApihost, flagAuth, flagInclude, flagExclude})
	if err != nil && len(output.Captured) == 0 {
		// Just an error, nothing in 'Captured'
		return err
	}
	// The output from "project/deploy" is not quite right for doctl even with branding, so fix up
	// what is in 'Captured'.  We do this even if there has been an error, because the output of
	// deploy is complex and the transcript is often needed to interpret the error.
	for index, value := range output.Captured {
		if strings.Contains(value, "Deploying project") {
			output.Captured[index] = strings.Replace(value, "Deploying project", "Deployed", 1)
		} else if strings.Contains(value, "Deployed actions") {
			output.Captured[index] = "Deployed functions ('doctl sbx fn get <funcName> --url' for URL):"
		}
	}
	if err == nil {
		// Normal error-free return
		return c.PrintServerlessTextOutput(output)
	}
	// When there is an error but also a transcript, display the transcript before return the error
	// This is "best effort" so we ignore any error returns from the print statement
	fmt.Fprintln(c.Out, strings.Join(output.Captured, "\n"))
	return err
}

// writeAFile is a thin wrapper around os.WriteFile designed to be replaced for testing.
var writeAFile = func(path string, contents []byte) error {
	return os.WriteFile(path, contents, 0664)
}

// doMkdir is a thin wrapper around os.Mkdir or os.MkdirAll designed to be replaced for testing.
var doMkdir = func(path string, parents bool) error {
	if parents {
		return os.MkdirAll(path, 0775)
	}
	return os.Mkdir(path, 0775)
}

// RunServerlessExtraGetMetadata supports the 'serverless get-metadata' command
func RunServerlessExtraGetMetadata(c *CmdConfig) error {
	adjustIncludeAndExclude(c)
	err := ensureOneArg(c)
	if err != nil {
		return err
	}
	r, _ := c.Doit.GetBool(c.NS, flagProjectReader)

	var output do.ServerlessOutput
	project := do.ServerlessProject{
		ProjectPath: c.Args[0],
	}
	if r {
		output, err = c.Serverless().ReadProject(&project, c.Args)
	} else {
		output, err = RunServerlessExec("get-metadata", c, []string{flagJSON, flagProjectReader}, []string{flagEnv, flagInclude, flagExclude})
	}
	if err != nil {
		return err
	}
	return c.PrintServerlessTextOutput(output)
}

// RunServerlessExtraWatch supports 'serverless watch'
// This is not the usual boiler-plate because the command is intended to be long-running in a separate window
func RunServerlessExtraWatch(c *CmdConfig) error {
	adjustIncludeAndExclude(c)
	err := ensureOneArg(c)
	if err != nil {
		return err
	}
	return RunServerlessExecStreaming("watch", c, []string{flagInsecure, flagVerboseBuild, flagVerboseZip, flagYarn, flagRemoteBuild},
		[]string{flagEnv, flagBuildEnv, flagApihost, flagAuth, flagInclude, flagExclude})
}

// prepareProjectArea prepares a disk area for receiving a project.  If the area exists and is not empty,
// the overwrite flag must be true else error.  On successful return, the area either did not pre-exist
// or has been removed.  Note: this function can be replaced for testing.
var prepareProjectArea = func(project string, overwrite bool) error {
	_, err := os.Stat(project)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	// project exists in the file system
	if !overwrite {
		return fmt.Errorf("%s already exists; use '--overwrite' to replace", project)
	}
	// overwrite was specified: it is permitted to remove what's there
	return os.RemoveAll(project)
}

// fileExtensionForRuntime maps runtimes to extensions in naming samples (other than typescript, which is handled specially)
// Considering that we use dynamic information in deciding what runtimes are supported, this is disappointingly static but we
// don't have a good way to make it dynamic at present.
func fileExtensionForRuntime(runtime string) string {
	switch runtime {
	case "nodejs":
		return "js"
	case "python":
		return "py"
	}
	// At present, for others, e.g. 'go' and 'php', runtime == extension
	return runtime
}

// generateSample generates a sample function in the sample package
func generateSample(kind string, config *do.ServerlessSpec, sample string, samplePackage string, ts bool) (string, error) {
	runtime := strings.Split(kind, ":")[0]
	var suffix string
	if ts {
		suffix = "ts"
	} else {
		suffix = fileExtensionForRuntime(runtime)
	}
	actionDir := filepath.Join(samplePackage, "hello")
	if err := doMkdir(actionDir, true); err != nil {
		return "", err
	}
	var file string
	if ts {
		srcDir := filepath.Join(actionDir, "src")
		if err := doMkdir(srcDir, false); err != nil {
			return "", err
		}
		file = filepath.Join(srcDir, "hello."+suffix)
	} else {
		file = filepath.Join(actionDir, "hello."+suffix)
	}
	if err := writeAFile(file, []byte(sample)); err != nil {
		return "", err
	}
	var sampPkg *do.ServerlessPackage
	for _, pkg := range config.Packages {
		if pkg.Name == "sample" {
			sampPkg = pkg
			break
		}
	}
	if sampPkg.Name != "sample" {
		return "", fmt.Errorf("could not find sample package in config (internal error)")
	}
	function := do.ServerlessFunction{
		Name:      "hello",
		Runtime:   kind,
		Web:       true,
		WebSecure: false,
	}
	sampPkg.Functions = []*do.ServerlessFunction{&function}
	return actionDir, nil
}

// languageToKindAndSample converts a user-specified language name to a runtime kind plus a sample.
// A third return value indicates that the language is typescript (special support in the sample
// generation).  Returns an error if the user requests an unsupported language.
func languageToKindAndSample(c *CmdConfig, language string) (string, string, bool, error) {
	language = strings.ToLower(language)
	if !isValidLanguage(c, language) {
		return "", "", false, fmt.Errorf("%s is not a supported language", language)
	}
	runtime := language
	ts := false
	switch language {
	case "ts", "typescript":
		ts = true
		runtime = "nodejs"
	case "js", "javascript":
		runtime = "nodejs"
	case "py":
		runtime = "python"
	case "golang":
		runtime = "go"
	}
	return runtime + ":default", samples[language], ts, nil
}

// isValidLanguage uses the languageKeywords table to decide if a language name is valid.
// If it appears to be valid, a runtime check is run.  Since we don't want to require
// connectivity to create a project, we accept the keyword if we can't contact the host
// and only reject it if we do contact the host and the host says there's no such runtime.
func isValidLanguage(c *CmdConfig, language string) bool {
	for runtime, kwds := range languageKeywords {
		for _, kwd := range kwds {
			if language == kwd {
				return validateRuntime(c, runtime)
			}
		}
	}
	return false
}

// configTemplate builds a minimal project configuration (project.yml) in memory.
func configTemplate() do.ServerlessSpec {
	config := do.ServerlessSpec{}
	defPkg := do.ServerlessPackage{Name: "sample"}
	config.Packages = []*do.ServerlessPackage{&defPkg}
	return config
}

// validateRuntime takes the name of a runtime and tries to contact the host to
// determine if that runtime exists.  It returns false only if it succeeds in
// contacting the host and is told the runtime does not exist.
func validateRuntime(c *CmdConfig, runtime string) bool {
	if runtime == "nodejs" {
		// As long as we are using a default language of 'javascript' we must necessarily assume
		// that the 'nodejs' runtime is valid.  No need to go through the overhead of checking.
		return true
	}
	sls := c.Serverless()
	err := sls.CheckServerlessStatus()
	if err != nil {
		return true // err in the permissive direction
	}
	creds, err := sls.ReadCredentials()
	if err != nil {
		return true // err in the permissive direction
	}
	info, err := sls.GetHostInfo(creds.APIHost)
	if err != nil {
		return true // err in the permissive direction
	}
	for validRuntime := range info.Runtimes {
		if runtime == validRuntime {
			return true
		}
	}
	return false
}

// adjustIncludeAndExclude deals with the fact that 'web' has special meaning to 'nim'.
// 1.  If the developer has a package called 'web' and wishes to include or exclude it, 'nim' will be confused unless a trailing
// slash is added to indicate that the intent is the package called 'web' and not 'web content.'
// 2.  Since projects may have a non-empty 'web' folder, 'nim' will want to deploy it unless '--exclude web' is provided.
// Note that the developer may already by using '--exclude', so this additional exclusion will be an append to the existing
// value.
func adjustIncludeAndExclude(c *CmdConfig) {
	includes, err := c.Doit.GetString(c.NS, flagInclude)
	if err == nil && includes != "" {
		includes = qualifyWebWithSlash(includes)
		c.Doit.Set(c.NS, flagInclude, includes)
	}
	excludes, err := c.Doit.GetString(c.NS, flagExclude)
	if err == nil && excludes != "" {
		excludes = qualifyWebWithSlash(excludes)
		excludes = excludes + "," + keywordWeb
		c.Doit.Set(c.NS, flagExclude, excludes)
	} else {
		c.Doit.Set(c.NS, flagExclude, keywordWeb)
	}
}

// qualifyWebWithSlash is a subroutine used by adjustIncludeAndExclude.  Given a comma-separated
// list of tokens, if any of those tokens are 'web', change that token to 'web/' and return the
// modified list.
func qualifyWebWithSlash(original string) string {
	tokens := strings.Split(original, ",")
	for i, token := range tokens {
		if token == "web" {
			tokens[i] = "web/"
		}
	}
	return strings.Join(tokens, ",")
}