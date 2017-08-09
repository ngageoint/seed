package constants

//TrueString string version of true boolean
const TrueString = "true"

//BuildCommand seed build subcommand
const BuildCommand = "build"

//RunCommand seed run subcommand
const RunCommand = "run"

//ListCommand seed list subcommand
const ListCommand = "list"

//SearchCommand seed search subcommand
const SearchCommand = "search"

//PublishCommand seed publish subcommand
const PublishCommand = "publish"

//ValidateCommand seed validate subcommand
const ValidateCommand = "validate"

//VersionCommand seed version subcommand
const VersionCommand = "version"

//JobDirectoryFlag defines the location of the seed spec and Dockerfile
const JobDirectoryFlag = "directory"

//ShortJobDirectoryFlag defines the shorthand location of the seed spec and Dockerfile
const ShortJobDirectoryFlag = "d"

//SettingFlag defines the SettingFlag
const SettingFlag = "setting"

//ShortSettingFlag defines the shorthand SettingFlag
const ShortSettingFlag = "e"

//MountFlag defines the MountFlag
const MountFlag = "mount"

//ShortMountFlag defines the shorthand MountFlag
const ShortMountFlag = "m"

//InputDataFlag defines the InputDataFlag
const InputDataFlag = "inputData"

//ShortInputDataFlag defines the shorthand input data flag
const ShortInputDataFlag = "i"

//JobOutputDirFlag defines the job output directory
const JobOutputDirFlag = "outDir"

//ShortJobOutputDirFlag defines the shorthand output directory
const ShortJobOutputDirFlag = "o"

//ShortImgNameFlag defines image name to run
const ShortImgNameFlag = "in"

//ImgNameFlag defines image name to run
const ImgNameFlag = "imageName"

//RmFlag defines if the docker image should be removed after docker run is executed
const RmFlag = "rm"

//SchemaFlag defines a schema file to validate seed against
const SchemaFlag = "schema"

//ShortSchemaFlag shorthand flag that defines schema file to validate seed against
const ShortSchemaFlag = "s"

//RegistryFlag defines registry
const RegistryFlag = "registry"

//ShortRegistryFlag shorthand flag that defines registry
const ShortRegistryFlag = "r"

//OrgFlag defines organization
const OrgFlag = "org"

//ShortOrgFlag shorthand flag that defines organization
const ShortOrgFlag = "o"

//FilterFlag defines filter
const FilterFlag = "filter"

//ShortFilterFlag shorthand flag that defines filter
const ShortFilterFlag= "f"

//UserFlag defines user
const UserFlag = "user"

//ShortUserFlag shorthand flag that defines user
const ShortUserFlag = "u"

//PassFlag defines password
const PassFlag = "password"

//ShortPassFlag shorthand flag that defines password
const ShortPassFlag = "p"

//ForcePublishFlag forces a publish - don't try to deconflict
const ForcePublishFlag = "f"

//PkgVersionMinor specifies to bump package minor version
const PkgVersionMinor = "p"

//PkgVersionMajor specifies to bump package major version
const PkgVersionMajor = "P"

//AlgVersionMinor specifies to bump algorithm minor version
const AlgVersionMinor = "a"

//AlgVersionMajor specifies to bump algorithm major version
const AlgVersionMajor = "A"

//SeedFileName defines the filename for the seed file
const SeedFileName = "seed.manifest.json"

//ResultsFileManifestName defines the filename for the results_manifest file
const ResultsFileManifestName = "results_manifest.json"

//DefaultRegistry defines the default registry address to use when searching for images
const DefaultRegistry = "https://hub.docker.com/v2"

//DefaultOrg defines the default organization to use when searching for images
const DefaultOrg = "geoint"

//SchemaType defines manfiest or metadata
type SchemaType int

const (
	//SchemaManifest manifest schema
	SchemaManifest SchemaType = iota

	//SchemaMetadata metadata schema
	SchemaMetadata
)
