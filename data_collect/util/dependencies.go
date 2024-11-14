package util

import (
	"fmt"
	"data_collection/models"
	"regexp"
	"sort"
	"strings"
)

const (
	library = "(?i)library\\('([A-Za-z]+(\\.[A-Za-z]+)+)', '([A-Za-z]+(\\.[A-Za-z]+)+):([A-Za-z]+(-[A-Za-z]+)+):([A-Za-z0-9]+(\\.[A-Za-z0-9]+)+)'\\)"
)

// Scan dependencies of one repo and check whether it is in Map
func CheckDependencies(dependencies []string, dependenciesMap, UnrelatedPackageMap, newDependenciesMap map[string]models.Dependency) (bool, map[string]models.Dependency, map[string]models.Dependency, []models.Dependency, []string) {
	hasAimPacks := false
	unHandledDependencies := make([]string, 0)
	// get dependencies elements info
	new_dependency := make([]models.Dependency, 0)
	for _, dependency := range dependencies {
		if strings.Contains(dependency, "versionRef") {
			continue
		}
		dependency = strings.Replace(dependency, "-", "", -1)
		info1 := getImplementationPatternDependInfo(dependency)
		info2 := getLibraryPatternDependInfo(dependency)
		info3 := getClassPathPatternDependInfo(dependency)
		info := findValidInfo(info1, info2, info3)
		if info.MethodName == "" && info.PackageName == "" {
			unHandledDependencies = append(unHandledDependencies, dependency)
		}
		// if already existed packages but not sure functions
		if _, ok := dependenciesMap[info.PackageName]; ok {
			info.Number += dependenciesMap[info.PackageName].Number
			dependenciesMap[info.PackageName] = info
			hasAimPacks = true
			fmt.Println(hasAimPacks, "depend ", info)
			// First neglect function name
			// same package but not same function, manually check!
			// if dependenciesMap[info.PackageName].MethodName != info.MethodName {
			// 	new_dependency = append(new_dependency, info)
			// }
		} else {
			if _, ok1 := UnrelatedPackageMap[info.PackageName]; !ok1 {
				// New package, manually check!
				if _, ok := newDependenciesMap[info.PackageName]; ok {
					info.Number += newDependenciesMap[info.PackageName].Number
					newDependenciesMap[info.PackageName] = info
				} else {
					newDependenciesMap[info.PackageName] = info
					new_dependency = append(new_dependency, info)
				}
			}
		}
	}
	return hasAimPacks, dependenciesMap, newDependenciesMap, new_dependency, unHandledDependencies
}

func getLibraryPatternDependInfo(dependency string) models.Dependency {
	//library('errorprone-gradle', 'net.ltgt.gradle:gradle-errorprone-plugin:2.0.2')
	library_pattern1 := regexp.MustCompile("(?i)library\\('[A-Za-z0-9]+', '([A-Za-z]+(\\.[A-Za-z]+)+):[A-Za-z]+:([A-Za-z0-9]+(\\.[A-Za-z0-9]+)+)'\\)")
	//library('androidxespresso.idling', 'androidx.test.espresso.idling', 'idlingconcurrent').versionRef('androidxespresso')
	library_pattern2 := regexp.MustCompile("(?i)library\\('([A-Za-z]+(\\.[A-Za-z]+)+)', '([A-Za-z]+(\\.[A-Za-z]+)+)', '[A-Za-z0-9]+'\\)\\.versionRef\\('[A-Za-z0-9]+'\\)")
	// library('daggerandroid', 'com.google.dagger', 'daggerandroid').versionRef('dagger')
	library_pattern3 := regexp.MustCompile("(?i)library\\('[A-Za-z0-9]+', '([A-Za-z]+(\\.[A-Za-z]+)+)', '[A-Za-z0-9]+'\\)\\.versionRef\\('[A-Za-z0-9]+'\\)")
	// library('androidxtest.ktx', 'androidx.test:corektx:1.5.0')
	library_pattern4 := regexp.MustCompile("(?i)library\\('([A-Za-z0-9]+(\\.[A-Za-z0-9]+)+)', '([A-Za-z0-9]+(\\.[A-Za-z0-9]+)+):[A-Za-z0-9]+:([0-9]+(\\.[0-9]+)+)'\\)")
	// library_pattern5 := regexp.MustCompile("(?i)library\\('[A-Za-z0-9]+', '[^']*'\\)")
	// // library("androidxviewpager", "androidx.viewpager2:viewpager2:$viewpager")
	// library_pattern6 := regexp.MustCompile("(?i)library\\(\"[A-Za-z0-9]+\", \"[A-Za-z0-9]+\\.[A-Za-z0-9]+:[A-Za-z0-9]+:\\$[A-Za-z]+\"\\)")

	info := models.Dependency{}
	info.Number = 1
	info1 := library_pattern1.FindStringSubmatch(dependency)
	if len(info1) > 4 {
		info.PackageName = info1[1]
		info.MethodName = info1[2]
		if len(info1) > 4 {
			info.Version = info1[3]
		}
		return info
	}
	info2 := library_pattern2.FindStringSubmatch(dependency)
	if len(info2) > 4 {
		// no version
		info.PackageName = info2[3]
		info.MethodName = info2[2]
		return info
	}
	info3 := library_pattern3.FindStringSubmatch(dependency)
	if len(info3) > 3 {
		info.PackageName = info3[1]
		info.MethodName = info3[2]
		return info
	}
	info4 := library_pattern4.FindStringSubmatch(dependency)
	if len(info4) > 6 {
		info.PackageName = info4[4]
		info.MethodName = info4[2]
		if len(info4) > 6 {
			info.Version = info4[5]
		}
		return info
	}
	i := strings.Split(dependency, "'")
	if len(i) > 4 {
		infos := strings.Split(i[3], ":")
		if len(infos) >= 3 {
			info.PackageName = infos[0]
			info.MethodName = infos[1]
			info.Version = infos[2]
			return info
		}
	}
	return info

}
func getImplementationPatternDependInfo(dependency string) models.Dependency {
	// implementation "com.google.android.exoplayer:exoplayercore:${exoPlayerVersion}"
	implementation_pattern := regexp.MustCompile("(?i)implementation \"([A-Za-z0-9]+(\\.[A-Za-z0-9]+)+):[A-Za-z0-9]+:\\$\\{[A-Za-z0-9]+\\}\"")
	info1 := implementation_pattern.FindStringSubmatch(dependency)
	info := models.Dependency{}
	info.Number = 1
	if len(info1) > 2 {
		info.PackageName = info1[0]
		info.MethodName = info1[1]
		if len(info1) > 3 {
			info.Version = info1[2]
		}
	}
	return info
}

func getClassPathPatternDependInfo(dependency string) models.Dependency {
	// classpath("com.google.android.gms:oss-licenses-plugin:0.10.6")
	classpath_pattern := regexp.MustCompile("(?i)classpath\\(\"([A-Za-z0-9]+(\\.[A-Za-z0-9]+)+):([A-Za-z]+(-[A-Za-z]+)+):([A-Za-z0-9]+(\\.[A-Za-z0-9]+)+)\"\\)")

	info1 := classpath_pattern.FindStringSubmatch(dependency)
	info := models.Dependency{}
	info.Number = 1
	if len(info1) > 2 {
		info.PackageName = info1[0]
		info.MethodName = info1[1]
		if len(info1) > 3 {
			info.Version = info1[2]
		}
	}
	return info
}

// find valid dependencies info from three diff patterns
func findValidInfo(info1, info2, info3 models.Dependency) models.Dependency {
	if checkDependInfoIsValid(info1) {
		return info1
	} else if checkDependInfoIsValid((info2)) {
		return info2
	} else if checkDependInfoIsValid(info3) {
		return info3
	} else {
		return models.Dependency{}
	}
}

// check whether dependency is valid
func checkDependInfoIsValid(info models.Dependency) bool {
	if info.PackageName == "" && info.MethodName == "" {
		return false
	}
	return true
}

// Sort dependencies by used number
func SortDependenciesByUsedNumber(dependencies []models.Dependency) []models.Dependency {
	sort.SliceStable(dependencies, func(i, j int) bool {
		return dependencies[i].Number > dependencies[j].Number
	})
	return dependencies
}
