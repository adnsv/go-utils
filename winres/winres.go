package winres

import (
	"github.com/adnsv/go-utils/git"
	"github.com/josephspurrier/goversioninfo"
)

func MakeFileVersion(gitver *git.VersionInfo) goversioninfo.FileVersion {
	return goversioninfo.FileVersion{
		Major: int(gitver.Semantic.Major),
		Minor: int(gitver.Semantic.Minor),
		Patch: int(gitver.Semantic.Patch),
		Build: int(gitver.AdditionalCommits),
	}
}

func NewVersionInfo(productver, filever *git.VersionInfo) *goversioninfo.VersionInfo {
	v := &goversioninfo.VersionInfo{}
	v.FileFlagsMask = "3F"
	v.FileFlags = "00"
	v.FileOS = "040004"
	v.FileType = "01"
	v.FileSubType = "00"
	v.VarFileInfo.Translation.LangID = goversioninfo.LangID(0x0409)
	v.VarFileInfo.Translation.CharsetID = goversioninfo.CharsetID(0x04B0)

	if productver != nil {
		v.FixedFileInfo.ProductVersion = MakeFileVersion(productver)
		v.StringFileInfo.ProductVersion = productver.Semantic.String()
	}
	if filever != nil {
		v.FixedFileInfo.FileVersion = MakeFileVersion(filever)
		v.StringFileInfo.FileVersion = filever.Semantic.String()
	}
	return v
}
