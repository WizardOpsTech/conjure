package source

// Source is the interface for accessing templates and bundles
type Source interface {
	ListTemplates() ([]TemplateInfo, error)

	ListBundles() ([]BundleInfo, error)

	GetTemplate(name, version string) (*TemplateContent, error)

	GetBundle(name, version string) (*BundleContent, error)

	GetTemplateVersions(name string) ([]string, error)

	GetBundleVersions(name string) ([]string, error)
}
