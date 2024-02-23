package api

type Package interface {
	Install() error
	String() string
}

type PackageInstaller interface {
	EnsurePackagerInstalled(homedir string) error
	UpgradePackages() error
	DevelopmentTools() Package
	ShellTools() Package
	Desktop() Package
}

func InstallPackages(packages []Package) error {
	for _, pkg := range packages {
		if err := pkg.Install(); err != nil {
			return err
		}
	}
	return nil
}
