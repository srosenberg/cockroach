package main

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

func main() {
	__antithesis_instrumentation__.Notify(41248)
	log.SetFlags(0)
	var tokenPath, fromRef, toRef, repository, gitCheckoutDir string
	var dryRun bool

	flag.StringVar(&tokenPath, "token", "", "Path to a file containing the GitHub token with repo:public_repo scope")
	flag.StringVar(&fromRef, "from", "", "From git ref")
	flag.StringVar(&toRef, "to", "", "To git ref")
	flag.StringVar(&repository, "repo", "", "Github \"owner/repo\" to work on. Default: cockroachdb/cockroach")
	flag.StringVar(&gitCheckoutDir, "dir", "", "Git checkout directory")
	flag.BoolVar(&dryRun, "dry-run", false, "Dry run")
	flag.Parse()

	allRequiredFlagsSet := true
	if tokenPath == "" {
		__antithesis_instrumentation__.Notify(41259)
		allRequiredFlagsSet = false
		log.Println("missing required --token-path argument/flag")
	} else {
		__antithesis_instrumentation__.Notify(41260)
	}
	__antithesis_instrumentation__.Notify(41249)
	if fromRef == "" {
		__antithesis_instrumentation__.Notify(41261)
		allRequiredFlagsSet = false
		log.Println("missing required --from argument/flag")
	} else {
		__antithesis_instrumentation__.Notify(41262)
	}
	__antithesis_instrumentation__.Notify(41250)
	if toRef == "" {
		__antithesis_instrumentation__.Notify(41263)
		allRequiredFlagsSet = false
		log.Println("missing required --to argument/flag")
	} else {
		__antithesis_instrumentation__.Notify(41264)
	}
	__antithesis_instrumentation__.Notify(41251)
	if gitCheckoutDir == "" {
		__antithesis_instrumentation__.Notify(41265)
		allRequiredFlagsSet = false
		log.Println("missing required --dir argument/flag")
	} else {
		__antithesis_instrumentation__.Notify(41266)
	}
	__antithesis_instrumentation__.Notify(41252)

	if !allRequiredFlagsSet {
		__antithesis_instrumentation__.Notify(41267)
		log.Fatal("Try running with `-h` for help\nPlease provide all required parameters")
	} else {
		__antithesis_instrumentation__.Notify(41268)
	}
	__antithesis_instrumentation__.Notify(41253)

	if repository == "" {
		__antithesis_instrumentation__.Notify(41269)
		repository = "cockroachdb/cockroach"
		log.Printf("No repository specified. Using %s\n", repository)
	} else {
		__antithesis_instrumentation__.Notify(41270)
	}
	__antithesis_instrumentation__.Notify(41254)

	if gitCheckoutDir != "" {
		__antithesis_instrumentation__.Notify(41271)
		if err := os.Chdir(gitCheckoutDir); err != nil {
			__antithesis_instrumentation__.Notify(41272)
			log.Fatalf("Cannot chdir to %s: %+v\n", gitCheckoutDir, err)
		} else {
			__antithesis_instrumentation__.Notify(41273)
		}
	} else {
		__antithesis_instrumentation__.Notify(41274)
	}
	__antithesis_instrumentation__.Notify(41255)
	token, err := readToken(tokenPath)
	if err != nil {
		__antithesis_instrumentation__.Notify(41275)
		log.Fatalf("Cannot read token from %s: %+v\n", tokenPath, err)
	} else {
		__antithesis_instrumentation__.Notify(41276)
	}
	__antithesis_instrumentation__.Notify(41256)
	commonBaseRef, err := getCommonBaseRef(fromRef, toRef)
	if err != nil {
		__antithesis_instrumentation__.Notify(41277)
		log.Fatalf("Cannot get common base ref: %+v\n", err)
	} else {
		__antithesis_instrumentation__.Notify(41278)
	}
	__antithesis_instrumentation__.Notify(41257)

	refList, err := getRefs(commonBaseRef, toRef)
	if err != nil {
		__antithesis_instrumentation__.Notify(41279)
		log.Fatalf("Cannot get refs: %+v\n", err)
	} else {
		__antithesis_instrumentation__.Notify(41280)
	}
	__antithesis_instrumentation__.Notify(41258)

	for _, ref := range refList {
		__antithesis_instrumentation__.Notify(41281)
		tag, err := getFirstTagContainingRef(ref)
		if err != nil {
			__antithesis_instrumentation__.Notify(41286)
			log.Fatalf("Error getting first tag containing ref %s: %+v\n", ref, err)
		} else {
			__antithesis_instrumentation__.Notify(41287)
		}
		__antithesis_instrumentation__.Notify(41282)
		if tag == "" {
			__antithesis_instrumentation__.Notify(41288)
			log.Printf("Ref %s has not yet appeared in a released version.", ref)
			continue
		} else {
			__antithesis_instrumentation__.Notify(41289)
		}
		__antithesis_instrumentation__.Notify(41283)
		prs, err := getPrNumbers(ref)
		if err != nil {
			__antithesis_instrumentation__.Notify(41290)
			log.Fatalf("Cannot find PR for ref %s: %+v\n", ref, err)
		} else {
			__antithesis_instrumentation__.Notify(41291)
		}
		__antithesis_instrumentation__.Notify(41284)
		if len(prs) == 0 {
			__antithesis_instrumentation__.Notify(41292)
			log.Printf("No PRs for ref %s. There should be at least one.", ref)
			continue
		} else {
			__antithesis_instrumentation__.Notify(41293)
		}
		__antithesis_instrumentation__.Notify(41285)
		for _, pr := range prs {
			__antithesis_instrumentation__.Notify(41294)
			log.Printf("Labeling PR#%s (ref %s) using git tag %s", pr, ref, tag)
			if dryRun {
				__antithesis_instrumentation__.Notify(41296)
				log.Println("DRY RUN: skipping labeling")
				continue
			} else {
				__antithesis_instrumentation__.Notify(41297)
			}
			__antithesis_instrumentation__.Notify(41295)
			if err := labelPR(http.DefaultClient, repository, token, pr, tag); err != nil {
				__antithesis_instrumentation__.Notify(41298)
				log.Fatalf("Failed on label creation for Pull Request %s: '%s'\n", pr, err)
			} else {
				__antithesis_instrumentation__.Notify(41299)
			}
		}
	}
}

func readToken(path string) (string, error) {
	__antithesis_instrumentation__.Notify(41300)
	token, err := ioutil.ReadFile(path)
	if err != nil {
		__antithesis_instrumentation__.Notify(41302)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(41303)
	}
	__antithesis_instrumentation__.Notify(41301)
	return strings.TrimSpace(string(token)), nil
}

func getCommonBaseRef(fromRef, toRef string) (string, error) {
	__antithesis_instrumentation__.Notify(41304)
	cmd := exec.Command("git", "merge-base", fromRef, toRef)
	out, err := cmd.Output()
	if err != nil {
		__antithesis_instrumentation__.Notify(41306)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(41307)
	}
	__antithesis_instrumentation__.Notify(41305)
	return strings.TrimSpace(string(out)), nil
}

func filterPullRequests(text string) []string {
	__antithesis_instrumentation__.Notify(41308)
	var shas []string
	matchMerge := regexp.MustCompile(`Merge (#|pull request)`)
	for _, line := range strings.Split(text, "\n") {
		__antithesis_instrumentation__.Notify(41310)
		if !matchMerge.MatchString(line) {
			__antithesis_instrumentation__.Notify(41312)
			continue
		} else {
			__antithesis_instrumentation__.Notify(41313)
		}
		__antithesis_instrumentation__.Notify(41311)
		sha := strings.Fields(line)[0]
		shas = append(shas, sha)
	}
	__antithesis_instrumentation__.Notify(41309)
	return shas
}

func getRefs(fromRef, toRef string) ([]string, error) {
	__antithesis_instrumentation__.Notify(41314)
	cmd := exec.Command("git", "log", "--merges", "--reverse", "--oneline",
		"--format=format:%h %s", "--ancestry-path", fmt.Sprintf("%s..%s", fromRef, toRef))
	out, err := cmd.Output()
	if err != nil {
		__antithesis_instrumentation__.Notify(41316)
		return []string{}, err
	} else {
		__antithesis_instrumentation__.Notify(41317)
	}
	__antithesis_instrumentation__.Notify(41315)
	return filterPullRequests(string(out)), nil
}

func matchVersion(text string) string {
	__antithesis_instrumentation__.Notify(41318)
	for _, line := range strings.Fields(text) {
		__antithesis_instrumentation__.Notify(41320)

		regVersion := regexp.MustCompile(`^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)`)
		if !regVersion.MatchString(line) {
			__antithesis_instrumentation__.Notify(41325)
			continue
		} else {
			__antithesis_instrumentation__.Notify(41326)
		}
		__antithesis_instrumentation__.Notify(41321)

		alpha00Regex := regexp.MustCompile(`v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)-alpha.00+$`)
		if alpha00Regex.MatchString(line) {
			__antithesis_instrumentation__.Notify(41327)
			continue
		} else {
			__antithesis_instrumentation__.Notify(41328)
		}
		__antithesis_instrumentation__.Notify(41322)

		alphaBetaRcRegex := regexp.MustCompile(`^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)-[-.0-9A-Za-z]+$`)
		if alphaBetaRcRegex.MatchString(line) {
			__antithesis_instrumentation__.Notify(41329)
			return line
		} else {
			__antithesis_instrumentation__.Notify(41330)
		}
		__antithesis_instrumentation__.Notify(41323)

		patchRegex := regexp.MustCompile(`^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.[1-9][0-9]*$`)
		if patchRegex.MatchString(line) {
			__antithesis_instrumentation__.Notify(41331)
			return line
		} else {
			__antithesis_instrumentation__.Notify(41332)
		}
		__antithesis_instrumentation__.Notify(41324)

		majorReleaseRegex := regexp.MustCompile(`^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.0$`)
		if majorReleaseRegex.MatchString(line) {
			__antithesis_instrumentation__.Notify(41333)
			return line
		} else {
			__antithesis_instrumentation__.Notify(41334)
		}
	}
	__antithesis_instrumentation__.Notify(41319)
	return ""
}

func getFirstTagContainingRef(ref string) (string, error) {
	__antithesis_instrumentation__.Notify(41335)
	cmd := exec.Command("git", "tag", "--contains", ref)
	out, err := cmd.Output()
	if err != nil {
		__antithesis_instrumentation__.Notify(41337)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(41338)
	}
	__antithesis_instrumentation__.Notify(41336)
	version := matchVersion(string(out))
	return version, nil
}

func extractPrNumbers(text string) []string {
	__antithesis_instrumentation__.Notify(41339)
	var numbers []string
	lines := strings.SplitN(text, "\n", 2)
	for _, prNumber := range strings.Fields(lines[0]) {
		__antithesis_instrumentation__.Notify(41341)
		if strings.HasPrefix(prNumber, "#") {
			__antithesis_instrumentation__.Notify(41342)
			numbers = append(numbers, strings.TrimPrefix(prNumber, "#"))
		} else {
			__antithesis_instrumentation__.Notify(41343)
		}
	}
	__antithesis_instrumentation__.Notify(41340)
	return numbers
}

func getPrNumbers(ref string) ([]string, error) {
	__antithesis_instrumentation__.Notify(41344)
	cmd := exec.Command("git", "show", "--oneline", "--format=format:%h %s", ref)
	out, err := cmd.Output()
	if err != nil {
		__antithesis_instrumentation__.Notify(41346)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(41347)
	}
	__antithesis_instrumentation__.Notify(41345)
	prs := extractPrNumbers(string(out))
	return prs, nil
}

func apiCall(client *http.Client, url string, token string, payload interface{}) error {
	__antithesis_instrumentation__.Notify(41348)
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		__antithesis_instrumentation__.Notify(41353)
		return err
	} else {
		__antithesis_instrumentation__.Notify(41354)
	}
	__antithesis_instrumentation__.Notify(41349)
	req, err := http.NewRequest("POST", url, bytes.NewReader(payloadJSON))
	if err != nil {
		__antithesis_instrumentation__.Notify(41355)
		return err
	} else {
		__antithesis_instrumentation__.Notify(41356)
	}
	__antithesis_instrumentation__.Notify(41350)
	req.Header.Set("Authorization", fmt.Sprintf("token %s", token))
	resp, err := client.Do(req)
	if err != nil {
		__antithesis_instrumentation__.Notify(41357)
		return err
	} else {
		__antithesis_instrumentation__.Notify(41358)
	}
	__antithesis_instrumentation__.Notify(41351)
	defer resp.Body.Close()

	if !(resp.StatusCode == http.StatusCreated || func() bool {
		__antithesis_instrumentation__.Notify(41359)
		return resp.StatusCode == http.StatusUnprocessableEntity == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(41360)
		return resp.
			StatusCode == http.StatusOK == true
	}() == true) {
		__antithesis_instrumentation__.Notify(41361)
		return fmt.Errorf("status code %d from %s", resp.StatusCode, url)
	} else {
		__antithesis_instrumentation__.Notify(41362)
	}
	__antithesis_instrumentation__.Notify(41352)
	return nil
}

func labelPR(client *http.Client, repository string, token string, pr string, tag string) error {
	__antithesis_instrumentation__.Notify(41363)
	label := fmt.Sprintf("earliest-release-%s", tag)
	payload := struct {
		Name  string `json:"name"`
		Color string `json:"color"`
	}{
		Name:  label,
		Color: "000000",
	}
	if err := apiCall(client, fmt.Sprintf("https://api.github.com/repos/%s/labels", repository), token,
		payload); err != nil {
		__antithesis_instrumentation__.Notify(41366)
		return err
	} else {
		__antithesis_instrumentation__.Notify(41367)
	}
	__antithesis_instrumentation__.Notify(41364)
	url := fmt.Sprintf("https://api.github.com/repos/%s/issues/%s/labels", repository, pr)
	labels := []string{label}
	if err := apiCall(client, url, token, labels); err != nil {
		__antithesis_instrumentation__.Notify(41368)
		return err
	} else {
		__antithesis_instrumentation__.Notify(41369)
	}
	__antithesis_instrumentation__.Notify(41365)
	log.Printf("Label %s added to PR %s\n", label, pr)
	return nil
}
