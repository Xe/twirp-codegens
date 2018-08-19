package main

import (
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/pkg/errors"
	"github.com/twitchtv/protogen/stringutils"
)

// deduceGenPkgName figures out the go package name to use for generated code.
// Will try to use the explicit go_package setting in a file (if set, must be
// consistent in all files). If no files have go_package set, then use the
// protobuf package name (must be consistent in all files)
func deduceGenPkgName(genFiles []*descriptor.FileDescriptorProto) (string, error) {
	var genPkgName string
	for _, f := range genFiles {
		name, explicit := goPackageName(f)
		if explicit {
			name = stringutils.CleanIdentifier(name)
			if genPkgName != "" && genPkgName != name {
				// Make sure they're all set consistently.
				return "", errors.Errorf("files have conflicting go_package settings, must be the same: %q and %q", genPkgName, name)
			}
			genPkgName = name
		}
	}
	if genPkgName != "" {
		return genPkgName, nil
	}

	// If there is no explicit setting, then check the implicit package name
	// (derived from the protobuf package name) of the files and make sure it's
	// consistent.
	for _, f := range genFiles {
		name, _ := goPackageName(f)
		name = stringutils.CleanIdentifier(name)
		if genPkgName != "" && genPkgName != name {
			return "", errors.Errorf("files have conflicting package names, must be the same or overridden with go_package: %q and %q", genPkgName, name)
		}
		genPkgName = name
	}

	// All the files have the same name, so we're good.
	return genPkgName, nil
}
