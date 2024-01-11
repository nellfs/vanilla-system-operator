package cmd

/*	License: GPLv3
	Authors:
		Mirko Brombin <send@mirko.pm>
		Vanilla OS Contributors <https://github.com/vanilla-os/>
	Copyright: 2023
	Description: VSO is a utility which allows you to perform maintenance
	tasks on your Vanilla OS installation.
*/

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vanilla-os/orchid/cmdr"
	"github.com/vanilla-os/vso/core"
	bolt "go.etcd.io/bbolt"
)

func NewWayCommand() []*cmdr.Command {
	// Root command
	cmd := cmdr.NewCommand(
		"android",
		vso.Trans("waydroid.description"),
		vso.Trans("waydroid.description"),
		nil,
	)

	// Subcommands
	cleanCmd := cmdr.NewCommand(
		"clean",
		vso.Trans("waydroid.clean.description"),
		vso.Trans("waydroid.clean.description"),
		wayClean)

	deleteCmd := cmdr.NewCommand(
		"delete",
		vso.Trans("waydroid.delete.description"),
		vso.Trans("waydroid.delete.description"),
		wayDelete,
	)

	infoCmd := cmdr.NewCommand(
		"info",
		vso.Trans("waydroid.info.description"),
		vso.Trans("waydroid.info.description"),
		wayInfo,
	)

	installCmd := cmdr.NewCommand(
		"install",
		vso.Trans("waydroid.install.description"),
		vso.Trans("waydroid.install.description"),
		wayInstall,
	)
	installCmd.WithBoolFlag(
		cmdr.NewBoolFlag(
			"local",
			"l",
			vso.Trans("waydroid.install.options.local.description"),
			false,
		),
	)
	installCmd.WithBoolFlag(
		cmdr.NewBoolFlag(
			"noconfirm",
			"y",
			vso.Trans("waydroid.install.options.noconfirm.description"),
			false,
		),
	)

	initCmd := cmdr.NewCommand(
		"init",
		vso.Trans("waydroid.init.description"),
		vso.Trans("waydroid.init.description"),
		wayInit,
	)
	initCmd.WithBoolFlag(
		cmdr.NewBoolFlag(
			"force",
			"f",
			vso.Trans("waydroid.init.options.force.description"),
			false,
		),
	)

	launchCmd := cmdr.NewCommand(
		"launch",
		vso.Trans("waydroid.launch.description"),
		vso.Trans("waydroid.launch.description"),
		wayLaunch,
	)

	launcherCmd := cmdr.NewCommand(
		"launcher",
		vso.Trans("waydroid.launcher.description"),
		vso.Trans("waydroid.launcher.description"),
		wayLauncher,
	)

	removeCmd := cmdr.NewCommand(
		"remove",
		vso.Trans("waydroid.remove.description"),
		vso.Trans("waydroid.remove.description"),
		wayRemove,
	)

	searchCmd := cmdr.NewCommand(
		"search",
		vso.Trans("waydroid.search.description"),
		vso.Trans("waydroid.search.description"),
		waySearch,
	)

	syncCmd := cmdr.NewCommand(
		"sync",
		vso.Trans("waydroid.sync.description"),
		vso.Trans("waydroid.sync.description"),
		waySync,
	)

	updateCmd := cmdr.NewCommand(
		"update",
		vso.Trans("waydroid.update.description"),
		vso.Trans("waydroid.update.description"),
		wayUpdate,
	)

	// Add subcommands to root
	cmd.AddCommand(cleanCmd)
	cmd.AddCommand(deleteCmd)
	cmd.AddCommand(infoCmd)
	cmd.AddCommand(installCmd)
	cmd.AddCommand(initCmd)
	cmd.AddCommand(launchCmd)
	cmd.AddCommand(launcherCmd)
	cmd.AddCommand(removeCmd)
	cmd.AddCommand(searchCmd)
	cmd.AddCommand(syncCmd)
	cmd.AddCommand(updateCmd)

	return []*cmdr.Command{cmd}
}

func wayClean(cmd *cobra.Command, args []string) error {
	cmdr.Info.Println(vso.Trans("waydroid.clean.info.index"))
	_, err := os.Stat(core.IndexCacheDir)
	if !os.IsNotExist(err) {
		err = os.RemoveAll(core.IndexCacheDir)
		if err != nil {
			cmdr.Error.Println(vso.Trans("waydroid.clean.error.index"))
			fmt.Println(err)
		}
	}
	cmdr.Info.Println(vso.Trans("waydroid.clean.info.apk"))
	_, err = os.Stat(core.APKCacheDir)
	if !os.IsNotExist(err) {
		err = os.RemoveAll(core.APKCacheDir)
		if err != nil {
			cmdr.Error.Println(vso.Trans("waydroid.clean.error.apk"))
		}
	}
	cmdr.Info.Println(vso.Trans("waydroid.clean.info.success"))
	return nil
}

func wayDelete(cmd *cobra.Command, args []string) error {
	if core.AskConfirmation(vso.Trans("waydroid.delete.confirmation"), false) {
		return core.WayDelete()
	}
	return fmt.Errorf(vso.Trans("waydroid.delete.cancelled"))
}

func wayInfo(cmd *cobra.Command, args []string) error {
	search := strings.Join(args, " ")
	matches, err := core.SearchIndex(search, vso.Trans("waydroid.downloadIndex"))
	if err != nil {
		return err
	}
	var app core.FdroidPackage
	if len(matches) > 1 {
		var options []string
		for _, match := range matches {
			options = append(options, fmt.Sprintf("%s - %s [%s]", match.Name, match.Summary, match.Repository.Name))
		}
		selection := core.PickOption(vso.Trans("waydroid.info.info.PackageSelection"), options, 1)
		app = matches[selection]
	} else {
		app = matches[0]
	}
	fmt.Printf(vso.Trans("waydroid.info.PackageName"), app.Name)
	fmt.Println()
	fmt.Printf(vso.Trans("waydroid.info.InternalName"), app.RDNSName)
	fmt.Println()
	fmt.Printf(vso.Trans("waydroid.info.Summary"), app.Summary)
	fmt.Println()
	fmt.Printf(vso.Trans("waydroid.info.Author"), app.Author)
	fmt.Println()
	fmt.Printf(vso.Trans("waydroid.info.License"), app.License)
	fmt.Println()
	fmt.Printf(vso.Trans("waydroid.info.Repository"), app.Repository.Name)
	fmt.Println()
	return nil
}

func wayInit(cmd *cobra.Command, args []string) error {
	force, _ := cmd.Flags().GetBool("force")

	if core.WayExists() && !force {
		cmdr.Error.Println(vso.Trans("waydroid.init.error.alreadyInitialized"))
		return nil
	}

	err := core.WayInit()
	if err != nil {
		return err
	}

	cmdr.Success.Println(vso.Trans("waydroid.init.info.initialized"))
	return nil
}

func wayInstallRemote(search string, noconfirm bool, noprompt bool) (string, core.FdroidPackage, error) {
	_, err := os.Stat(core.APKCacheDir)
	if os.IsNotExist(err) {
		err := os.MkdirAll(core.APKCacheDir, 0755)
		if err != nil {
			return "", core.FdroidPackage{}, err
		}
	}

	matches, err := core.SearchIndex(search, vso.Trans("waydroid.downloadIndex"))
	if err != nil {
		return "", core.FdroidPackage{}, err
	}
	var match core.FdroidPackage
	if len(matches) <= 0 {
		fmt.Println("no match")
		return "", core.FdroidPackage{}, &core.NoMatchError{Search: search}
	} else if noprompt {
		useFirst := true
		for _, pkg := range matches {
			if strings.EqualFold(pkg.Name, search) {
				match = pkg
				useFirst = false
			}
		}
		if useFirst {
			match = matches[0]
		}
	} else if len(matches) > 1 {
		var options []string
		for _, match := range matches {
			options = append(options, fmt.Sprintf("%s - %s [%s]", match.Name, match.Summary, match.Repository.Name))
		}
		selection := core.PickOption(vso.Trans("waydroid.install.info.PackageSelection"), options, 1)
		match = matches[selection]
	} else {
		match = matches[0]
	}

	if noconfirm {
		cmdr.Info.Printfln(vso.Trans("waydroid.install.info.DownloadingPackage"), fmt.Sprintf(match.Repository.PackageURL, match.RDNSName))
		apk, err := core.FetchPackage(match)
		return apk, match, err
	} else if core.AskConfirmation(fmt.Sprintf(vso.Trans("waydroid.install.info.ConfirmInstall"), match.Name), true) {
		cmdr.Info.Printfln(vso.Trans("waydroid.install.info.DownloadingPackage"), fmt.Sprintf(match.Repository.PackageURL, match.RDNSName))
		apk, err := core.FetchPackage(match)
		return apk, match, err
	}
	return "", core.FdroidPackage{}, &core.InstallDeclined{}
}

func wayInstall(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		cmdr.Error.Println(vso.Trans("waydroid.error.noArguments"))
		return nil
	}
	localFlag, _ := cmd.Flags().GetBool("local")
	noconfirm, _ := cmd.Flags().GetBool("noconfirm")

	var err error
	var apk string
	var pkg core.FdroidPackage
	if !localFlag {
		apk, pkg, err = wayInstallRemote(strings.Join(args, " "), noconfirm, false)
		if err != nil {
			var NoMatchError *core.NoMatchError
			var PackageInCache *core.PackageInCache
			var InstallDeclined *core.InstallDeclined
			if errors.As(err, &NoMatchError) {
				cmdr.Error.Printfln(vso.Trans("waydroid.install.error.NotFound"), strings.Join(args, " "))
				return nil
			} else if errors.As(err, &PackageInCache) {
				cmdr.Info.Println(vso.Trans("waydroid.install.info.PackageInCache"))
			} else if errors.As(err, &InstallDeclined) {
				cmdr.Error.Println(vso.Trans("waydroid.install.error.InstallCancelled"))
				return nil
			} else {
				return err
			}
		}
	} else {
		fileName := strings.Split(args[0], "/")
		pkg = core.FdroidPackage{
			Name:       strings.ReplaceAll(fileName[len(fileName)-1], ".apk", ""),
			RDNSName:   strings.ReplaceAll(fileName[len(fileName)-1], ".apk", ""),
			Summary:    "",
			Author:     "",
			Source:     "",
			Repository: core.FdroidRepo{Name: "local"},
			Versions:   nil,
		}
		apk = args[0]
	}

	way, err := core.GetWay()
	if err != nil {
		return err
	}
	finalArgs := []string{"ewaydroid", "app", "install", apk}
	err = core.WayPutAppIntoDatabase(pkg, nil)
	if err != nil {
		return err
	}
	_, err = way.Exec(false, finalArgs...)
	return err
}

func wayLaunch(cmd *cobra.Command, args []string) error {
	way, err := core.GetWay()
	if err != nil {
		return err
	}

	finalArgs := []string{"ewaydroid", "app", "launch", args[0]}
	_, err = way.Exec(false, finalArgs...)
	return err
}

func wayLauncher(cmd *cobra.Command, args []string) error {
	way, err := core.GetWay()
	if err != nil {
		return err
	}

	finalArgs := []string{"ewaydroid", "show-full-ui"}
	_, err = way.Exec(false, finalArgs...)
	return err
}

func wayRemove(cmd *cobra.Command, args []string) error {
	way, err := core.GetWay()
	if err != nil {
		return err
	}
	search := strings.Join(args, " ")
	var matches [][]string
	var rem []string
	db, err := core.GetWayDatabase()
	if err != nil {
		return err
	}
	err = db.View(func(tx *bolt.Tx) error {
		apps := tx.Bucket([]byte("Apps"))
		if apps == nil {
			return &core.BucketNotFoundError{
				BucketName: "Apps",
			}
		}
		cursor := apps.Cursor()
		for key, value := cursor.First(); key != nil; key, value = cursor.Next() {
			var pkg core.FdroidPackage
			err := json.Unmarshal(value, &pkg)
			if err != nil {
				return err
			}
			if strings.Contains(strings.ToLower(pkg.Name), strings.ToLower(search)) || strings.Contains(strings.ToLower(pkg.RDNSName), strings.ToLower(search)) {
				matches = append(matches, []string{pkg.Name, pkg.RDNSName})
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	//defer db.Close()
	if len(matches) == 1 {
		rem = matches[0]
	} else if len(matches) > 1 {
		var options []string
		for _, match := range matches {
			options = append(options, fmt.Sprintf("%s (%s)", match[0], match[1]))
		}
		selection := core.PickOption(vso.Trans("waydroid.remove.info.PackageSelection"), options, 1)
		rem = matches[selection]
	} else {
		cmdr.Error.Printfln(vso.Trans("waydroid.remove.error.NoMatches"), strings.Join(args, " "))
		return nil
	}

	if !core.AskConfirmation(fmt.Sprintf(vso.Trans("waydroid.remove.info.ConfirmRemove"), fmt.Sprintf("%s (%s)", rem[0], rem[1])), true) {
		cmdr.Error.Println(vso.Trans("waydroid.remove.error.RemoveCancelled"))
		return nil
	}
	cmdr.Info.Printfln(vso.Trans("waydroid.remove.info.RemovePackage"), fmt.Sprintf("%s (%s)", rem[0], rem[1]))
	finalArgs := []string{"ewaydroid", "app", "remove", rem[1]}
	err = core.WayRemoveAppFromDatabase(rem[1], db)
	fmt.Println("here")
	if err != nil {
		return err
	}
	_, err = way.Exec(false, finalArgs...)
	return err
}

func waySearch(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		cmdr.Error.Println(vso.Trans("waydroid.error.noArguments"))
		return nil
	}

	search := strings.Join(args, " ") // Can only search for one thing at once, so might as well merge everything as one term
	matches, err := core.SearchIndex(search, vso.Trans("waydroid.downloadIndex"))
	if err != nil {
		return err
	}

	table := core.CreateVsoTable(os.Stdout)
	table.SetHeader([]string{"App", "Id", "Repository"})
	for _, match := range matches {
		table.Append([]string{match.Name + " - " + match.Summary, match.RDNSName, match.Repository.Name})
	}
	table.Render()

	return nil
}

func waySync(cmd *cobra.Command, args []string) error {
	err := core.GetRepos()
	if err != nil {
		return err
	}
	err = core.SyncIndex(true, vso.Trans("waydroid.downloadIndex"))
	return err
}

func wayUpdate(cmd *cobra.Command, args []string) error {
	db, err := core.GetWayDatabase()
	if err != nil {
		return err
	}
	var updates []core.FdroidPackage
	err = db.View(func(tx *bolt.Tx) error {
		apps := tx.Bucket([]byte("Apps"))
		if apps == nil {
			return &core.BucketNotFoundError{
				BucketName: "Apps",
			}
		}
		cursor := apps.Cursor()
		for key, value := cursor.First(); key != nil; key, value = cursor.Next() {
			var pkg core.FdroidPackage
			err := json.Unmarshal(value, &pkg)
			if err != nil {
				return err
			}
			latestVersion, err := core.GetPackageVersion(pkg)
			if err != nil {
				cmdr.Error.Printfln(vso.Trans("waydroid.update.error.FailGetVersion"), fmt.Sprintf("%s (%s)", pkg.Name, pkg.RDNSName))
				continue
			}
			latestVersionInt, err := strconv.ParseInt(latestVersion, 10, 0)
			if err != nil {
				continue
			}
			if !strings.Contains(pkg.Repository.Name, "local") && pkg.InstalledVersionCode < int(latestVersionInt) {
				updates = append(updates, pkg)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	if len(updates) == 0 {
		defer db.Close()
		cmdr.Info.Println(vso.Trans("waydroid.update.info.NoUpdates"))
		return nil
	}

	way, err := core.GetWay()
	if err != nil {
		return err
	}
	var PackageInCache *core.PackageInCache
	for _, update := range updates {
		apk, _, err := wayInstallRemote(update.RDNSName, true, true)
		if errors.As(err, &PackageInCache) {
			cmdr.Info.Println(vso.Trans("waydroid.install.info.PackageInCache"))
		} else if err != nil {
			cmdr.Error.Printfln(vso.Trans("waydroid.update.error.FailUpdatePackageDownload"), fmt.Sprintf("%s (%s)", update.Name, update.RDNSName))
			continue
		}
		finalArgs := []string{"ewaydroid", "app", "install", apk}
		latestVersion, err := core.GetPackageVersion(update)
		if err != nil {
			cmdr.Error.Printfln(vso.Trans("waydroid.update.error.FailGetVersion"), fmt.Sprintf("%s (%s)", update.Name, update.RDNSName))
			continue
		}
		latestVersionInt, err := strconv.ParseInt(latestVersion, 10, 0)
		if err != nil {
			continue
		}
		update.InstalledVersionCode = int(latestVersionInt)
		err = core.WayPutAppIntoDatabase(update, db)
		if err != nil {
			cmdr.Error.Printfln(vso.Trans("waydroid.update.error.FailUpdatePackageDatabase"), fmt.Sprintf("%s (%s)", update.Name, update.RDNSName))
			continue
		}
		_, err = way.Exec(false, finalArgs...)
		if err != nil {
			cmdr.Error.Printfln(vso.Trans("waydroid.update.error.FailUpdatePackageInstall"), fmt.Sprintf("%s (%s)", update.Name, update.RDNSName))
			continue
		}
	}
	defer db.Close()
	cmdr.Info.Println(vso.Trans("waydroid.update.finished"))
	return nil
}
