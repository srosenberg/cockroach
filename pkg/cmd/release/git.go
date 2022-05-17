package main

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
)

type releaseInfo struct {
	prevReleaseVersion string
	nextReleaseVersion string
	buildInfo          buildInfo

	candidateCommits []string

	releaseSeries string
}

func findNextVersion(releaseSeries string) (string, error) {
	__antithesis_instrumentation__.Notify(42615)
	prevReleaseVersion, err := findPreviousRelease(releaseSeries)
	if err != nil {
		__antithesis_instrumentation__.Notify(42618)
		return "", fmt.Errorf("cannot find previous release: %w", err)
	} else {
		__antithesis_instrumentation__.Notify(42619)
	}
	__antithesis_instrumentation__.Notify(42616)
	nextReleaseVersion, err := bumpVersion(prevReleaseVersion)
	if err != nil {
		__antithesis_instrumentation__.Notify(42620)
		return "", fmt.Errorf("cannot bump version: %w", err)
	} else {
		__antithesis_instrumentation__.Notify(42621)
	}
	__antithesis_instrumentation__.Notify(42617)
	return nextReleaseVersion, nil
}

func findNextRelease(releaseSeries string) (releaseInfo, error) {
	__antithesis_instrumentation__.Notify(42622)
	prevReleaseVersion, err := findPreviousRelease(releaseSeries)
	if err != nil {
		__antithesis_instrumentation__.Notify(42629)
		return releaseInfo{}, fmt.Errorf("cannot find previous release: %w", err)
	} else {
		__antithesis_instrumentation__.Notify(42630)
	}
	__antithesis_instrumentation__.Notify(42623)
	nextReleaseVersion, err := bumpVersion(prevReleaseVersion)
	if err != nil {
		__antithesis_instrumentation__.Notify(42631)
		return releaseInfo{}, fmt.Errorf("cannot bump version: %w", err)
	} else {
		__antithesis_instrumentation__.Notify(42632)
	}
	__antithesis_instrumentation__.Notify(42624)
	candidateCommits, err := findCandidateCommits(prevReleaseVersion, releaseSeries)
	if err != nil {
		__antithesis_instrumentation__.Notify(42633)
		return releaseInfo{}, fmt.Errorf("cannot find candidate commits: %w", err)
	} else {
		__antithesis_instrumentation__.Notify(42634)
	}
	__antithesis_instrumentation__.Notify(42625)
	info, err := findHealthyBuild(candidateCommits)
	if err != nil {
		__antithesis_instrumentation__.Notify(42635)
		return releaseInfo{}, fmt.Errorf("cannot find healthy build: %w", err)
	} else {
		__antithesis_instrumentation__.Notify(42636)
	}
	__antithesis_instrumentation__.Notify(42626)
	releasedVersions, err := getVersionsContainingRef(info.SHA)
	if err != nil {
		__antithesis_instrumentation__.Notify(42637)
		return releaseInfo{}, fmt.Errorf("cannot check if the candidate sha was released: %w", err)
	} else {
		__antithesis_instrumentation__.Notify(42638)
	}
	__antithesis_instrumentation__.Notify(42627)
	if len(releasedVersions) > 0 {
		__antithesis_instrumentation__.Notify(42639)
		return releaseInfo{}, fmt.Errorf("%s has been already released as a part of the following tags: %s",
			info.SHA, strings.Join(releasedVersions, ", "))
	} else {
		__antithesis_instrumentation__.Notify(42640)
	}
	__antithesis_instrumentation__.Notify(42628)
	return releaseInfo{
		prevReleaseVersion: prevReleaseVersion,
		nextReleaseVersion: nextReleaseVersion,
		buildInfo:          info,
		candidateCommits:   candidateCommits,
		releaseSeries:      releaseSeries,
	}, nil
}

func getVersionsContainingRef(ref string) ([]string, error) {
	__antithesis_instrumentation__.Notify(42641)
	cmd := exec.Command("git", "tag", "--contains", ref)
	out, err := cmd.Output()
	if err != nil {
		__antithesis_instrumentation__.Notify(42644)
		return []string{}, fmt.Errorf("cannot list tags containing %s: %w", ref, err)
	} else {
		__antithesis_instrumentation__.Notify(42645)
	}
	__antithesis_instrumentation__.Notify(42642)
	var versions []string
	for _, v := range findVersions(string(out)) {
		__antithesis_instrumentation__.Notify(42646)
		versions = append(versions, v.Original())
	}
	__antithesis_instrumentation__.Notify(42643)
	return versions, nil
}

func findVersions(text string) []*semver.Version {
	__antithesis_instrumentation__.Notify(42647)
	var versions []*semver.Version
	for _, line := range strings.Split(text, "\n") {
		__antithesis_instrumentation__.Notify(42649)
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" {
			__antithesis_instrumentation__.Notify(42653)
			continue
		} else {
			__antithesis_instrumentation__.Notify(42654)
		}
		__antithesis_instrumentation__.Notify(42650)

		if strings.Contains(trimmedLine, "-alpha.0000") {
			__antithesis_instrumentation__.Notify(42655)
			continue
		} else {
			__antithesis_instrumentation__.Notify(42656)
		}
		__antithesis_instrumentation__.Notify(42651)
		version, err := semver.NewVersion(trimmedLine)
		if err != nil {
			__antithesis_instrumentation__.Notify(42657)
			fmt.Printf("WARNING: cannot parse version '%s'\n", trimmedLine)
			continue
		} else {
			__antithesis_instrumentation__.Notify(42658)
		}
		__antithesis_instrumentation__.Notify(42652)
		versions = append(versions, version)
	}
	__antithesis_instrumentation__.Notify(42648)
	return versions
}

func findPreviousRelease(releaseSeries string) (string, error) {
	__antithesis_instrumentation__.Notify(42659)

	cmd := exec.Command("git", "tag", "--list", fmt.Sprintf("v%s.*", releaseSeries))
	output, err := cmd.Output()
	if err != nil {
		__antithesis_instrumentation__.Notify(42662)
		return "", fmt.Errorf("cannot get version from tags: %w", err)
	} else {
		__antithesis_instrumentation__.Notify(42663)
	}
	__antithesis_instrumentation__.Notify(42660)
	versions := findVersions(string(output))
	if len(versions) == 0 {
		__antithesis_instrumentation__.Notify(42664)
		return "", fmt.Errorf("zero versions found")
	} else {
		__antithesis_instrumentation__.Notify(42665)
	}
	__antithesis_instrumentation__.Notify(42661)
	sort.Sort(semver.Collection(versions))
	return versions[len(versions)-1].Original(), nil
}

func bumpVersion(version string) (string, error) {
	__antithesis_instrumentation__.Notify(42666)
	semanticVersion, err := semver.NewVersion(version)
	if err != nil {
		__antithesis_instrumentation__.Notify(42668)
		return "", fmt.Errorf("cannot parse version: %w", err)
	} else {
		__antithesis_instrumentation__.Notify(42669)
	}
	__antithesis_instrumentation__.Notify(42667)
	nextVersion := semanticVersion.IncPatch()
	return nextVersion.Original(), nil
}

func filterPullRequests(text string) []string {
	__antithesis_instrumentation__.Notify(42670)
	var shas []string
	matchMerge := regexp.MustCompile(`Merge (#|pull request)`)
	for _, line := range strings.Split(text, "\n") {
		__antithesis_instrumentation__.Notify(42672)
		if !matchMerge.MatchString(line) {
			__antithesis_instrumentation__.Notify(42674)
			continue
		} else {
			__antithesis_instrumentation__.Notify(42675)
		}
		__antithesis_instrumentation__.Notify(42673)
		sha := strings.Fields(line)[0]
		shas = append(shas, sha)
	}
	__antithesis_instrumentation__.Notify(42671)
	return shas
}

func getMergeCommits(fromRef, toRef string) ([]string, error) {
	__antithesis_instrumentation__.Notify(42676)
	cmd := exec.Command("git", "log", "--merges", "--format=format:%H %s", "--ancestry-path",
		fmt.Sprintf("%s..%s", fromRef, toRef))
	out, err := cmd.Output()
	if err != nil {
		__antithesis_instrumentation__.Notify(42678)
		return []string{}, fmt.Errorf("cannot read git log output: %w", err)
	} else {
		__antithesis_instrumentation__.Notify(42679)
	}
	__antithesis_instrumentation__.Notify(42677)
	return filterPullRequests(string(out)), nil
}

func getCommonBaseRef(fromRef, toRef string) (string, error) {
	__antithesis_instrumentation__.Notify(42680)
	cmd := exec.Command("git", "merge-base", fromRef, toRef)
	out, err := cmd.Output()
	if err != nil {
		__antithesis_instrumentation__.Notify(42682)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(42683)
	}
	__antithesis_instrumentation__.Notify(42681)
	return strings.TrimSpace(string(out)), nil
}

func findCandidateCommits(prevRelease string, releaseSeries string) ([]string, error) {
	__antithesis_instrumentation__.Notify(42684)
	releaseBranch := fmt.Sprintf("origin/release-%s", releaseSeries)
	commonBaseRef, err := getCommonBaseRef(prevRelease, releaseBranch)
	if err != nil {
		__antithesis_instrumentation__.Notify(42687)
		return []string{}, fmt.Errorf("cannot find common base ref: %w", err)
	} else {
		__antithesis_instrumentation__.Notify(42688)
	}
	__antithesis_instrumentation__.Notify(42685)
	refs, err := getMergeCommits(commonBaseRef, releaseBranch)
	if err != nil {
		__antithesis_instrumentation__.Notify(42689)
		return []string{}, fmt.Errorf("cannot get merge commits: %w", err)
	} else {
		__antithesis_instrumentation__.Notify(42690)
	}
	__antithesis_instrumentation__.Notify(42686)
	return refs, nil
}

func findHealthyBuild(potentialRefs []string) (buildInfo, error) {
	__antithesis_instrumentation__.Notify(42691)
	for _, ref := range potentialRefs {
		__antithesis_instrumentation__.Notify(42693)
		fmt.Println("Fetching release qualification metadata for", ref)
		meta, err := getBuildInfo(context.Background(), pickSHAFlags.qualifyBucket,
			fmt.Sprintf("%s/%s.json", pickSHAFlags.qualifyObjectPrefix, ref))
		if err != nil {
			__antithesis_instrumentation__.Notify(42695)

			fmt.Println("no metadata qualification for", ref, err)
			continue
		} else {
			__antithesis_instrumentation__.Notify(42696)
		}
		__antithesis_instrumentation__.Notify(42694)
		return meta, nil
	}
	__antithesis_instrumentation__.Notify(42692)
	return buildInfo{}, fmt.Errorf("no ref found")
}
