// Copyright 2024 The Cockroach Authors.
//
// Use of this software is governed by the CockroachDB Software License
// included in the /LICENSE file.

package mixedversion

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/roachtestutil/clusterupgrade"
	"github.com/cockroachdb/cockroach/pkg/testutils/release"
	"github.com/cockroachdb/version"
	"github.com/stretchr/testify/require"
)

// controlledSource is a rand.Source that returns predetermined values
// to control the outcomes of prng.Intn(2) calls in chooseSkips.
//
// For Go's math/rand, Intn(2) evaluates as:
//
//	Int31n(2) -> Int31() & 1 -> int32(source.Int63() >> 32) & 1
//
// So bit 32 of the Int63 return value determines the Intn(2) result.
type controlledSource struct {
	values []int64
	idx    int
}

func (s *controlledSource) Int63() int64 {
	if s.idx >= len(s.values) {
		return 0
	}
	v := s.values[s.idx]
	s.idx++
	return v
}

func (s *controlledSource) Seed(int64) {}

// bitsToSourceValues converts a bitmask into a sequence of Int63
// values such that sequential prng.Intn(2) calls yield the
// corresponding bits. Bit i of mask controls call i.
func bitsToSourceValues(mask, n int) []int64 {
	values := make([]int64, n)
	for i := range n {
		if mask&(1<<i) != 0 {
			values[i] = 1 << 32 // bit 32 set -> Intn(2) == 1
		}
		// else: 0 -> Intn(2) == 0
	}
	return values
}

// pathToString returns a canonical string representation of an upgrade path.
func pathToString(path []*clusterupgrade.Version) string {
	parts := make([]string, len(path))
	for i, v := range path {
		parts[i] = v.String()
	}
	return strings.Join(parts, " -> ")
}

// isDirectPredecessor returns true if fromSeries is the direct
// predecessor of toSeries according to the release data.
func isDirectPredecessor(fromSeries, toSeries string) bool {
	fromV := version.MustParse("v" + fromSeries + ".0")
	toV := version.MustParse("v" + toSeries + ".0")
	if !toV.AtLeast(fromV) {
		return false
	}
	n, err := release.MajorReleasesBetween(&fromV, &toV)
	return err == nil && n == 1
}

// isTransitivePredecessor returns true if fromSeries is a
// (possibly transitive) predecessor of toSeries. Directional:
// fromSeries must be strictly older than toSeries.
func isTransitivePredecessor(fromSeries, toSeries string) bool {
	fromV := version.MustParse("v" + fromSeries + ".0")
	toV := version.MustParse("v" + toSeries + ".0")
	if !toV.AtLeast(fromV) || fromSeries == toSeries {
		return false
	}
	n, err := release.MajorReleasesBetween(&fromV, &toV)
	return err == nil && n > 0
}

// verifyPathProperties checks invariant properties that must hold for
// every upgrade path produced by chooseUpgradePath.
func verifyPathProperties(
	t *testing.T,
	path []*clusterupgrade.Version,
	numUpgrades int,
	skipVersions bool,
	mbv, msv *clusterupgrade.Version,
) {
	t.Helper()
	currentVersion := clusterupgrade.CurrentVersion()

	// Property 1: Path is non-empty.
	require.NotEmpty(t, path, "path must be non-empty")

	// Property 2: Path ends with current version.
	require.True(t, path[len(path)-1].Equal(currentVersion),
		"path must end with current version %s, got %s", currentVersion, path[len(path)-1])

	// Property 3: Path length is at most numUpgrades + 1.
	// (numUpgrades predecessor steps + current version appended.)
	require.LessOrEqual(t, len(path), numUpgrades+1,
		"path has %d versions but numUpgrades=%d (max %d)", len(path), numUpgrades, numUpgrades+1)

	// Property 4: Path is strictly ordered (older to newer).
	for i := 1; i < len(path); i++ {
		require.True(t, path[i].AtLeast(path[i-1]),
			"path not ordered at position %d: %s should be >= %s", i, path[i], path[i-1])
		require.False(t, path[i].Equal(path[i-1]),
			"path has duplicate at position %d: %s", i, path[i])
	}

	// Property 5: All versions >= mbv.
	if mbv != nil {
		for _, v := range path {
			require.True(t, v.AtLeast(mbv),
				"version %s is below minimum bootstrap version %s", v, mbv)
		}
	}

	// Property 6: Each step is a valid upgrade (direct or transitive predecessor).
	for i := 1; i < len(path); i++ {
		fromSeries := path[i-1].Series()
		toSeries := path[i].Series()
		if skipVersions {
			require.True(t, isTransitivePredecessor(fromSeries, toSeries),
				"step %s -> %s: %s is not a transitive predecessor of %s",
				path[i-1], path[i], fromSeries, toSeries)
		} else {
			require.True(t, isDirectPredecessor(fromSeries, toSeries),
				"step %s -> %s: %s is not the direct predecessor of %s",
				path[i-1], path[i], fromSeries, toSeries)
		}
	}

	// Property 7: MSV-series predecessor is never skipped.
	// If skipVersions=true, the MSV-series predecessor must not be
	// removed by chooseSkips. However, the MSV series may simply not
	// appear in the path because numUpgrades truncated the chain before
	// reaching it. We only check when the MSV series falls between the
	// path's oldest version and the current version (exclusive).
	if msv != nil && skipVersions && len(path) > 1 {
		msvSeries := msv.Series()
		startSeries := path[0].Series()
		// The MSV series is within the path's range only if the path
		// starts at or before the MSV series.
		msvInRange := startSeries == msvSeries ||
			isTransitivePredecessor(startSeries, msvSeries)
		if msvInRange && isTransitivePredecessor(msvSeries, currentVersion.Series()) {
			hasMsvSeries := false
			for _, v := range path[:len(path)-1] {
				if v.Series() == msvSeries {
					hasMsvSeries = true
					break
				}
			}
			require.True(t, hasMsvSeries,
				"MSV series %s was skipped in path %s", msvSeries, pathToString(path))
		}
	}
}

// Test_chooseUpgradePath_exhaustive is a property-based test that
// exhaustively enumerates all possible upgrade paths produced by
// chooseUpgradePath.
//
// For each (finalVersion, numUpgrades, skipVersions) configuration,
// the test uses latestPredecessor (deterministic patch selection) and
// a controlled PRNG source to enumerate all 2^n skip combinations,
// where n is the number of predecessors in the chain. Every resulting
// path is verified against a set of invariant properties.
//
// The total number of iterations is bounded by:
//
//	sum over configs of 2^(predecessors in config)
//
// which is small enough to run exhaustively (typically < 10K paths).
func Test_chooseUpgradePath_exhaustive(t *testing.T) {
	defer setDefaultVersions()()

	// Test across multiple final versions that exercise different
	// skip-upgrade eligibility rules.
	finalVersions := []string{
		"v24.1.0", // no skip support
		"v24.3.0", // first version supporting skip upgrades (24.x)
		"v25.1.0", // 25.x non-skip version (ordinal 1)
		"v25.2.0", // 25.x skip version (ordinal 2)
		"v25.3.0", // 25.x non-skip version (ordinal 3)
	}

	var totalPaths int

	for _, fv := range finalVersions {
		t.Run(fv, func(t *testing.T) {
			defer withTestBuildVersion(fv)()

			currentV := clusterupgrade.CurrentVersion()
			maxUpgrades, err := release.MajorReleasesBetween(
				&currentV.Version, &minimumSupported.Version)
			if err != nil {
				t.Skipf("cannot compute upgrades for %s: %v", fv, err)
				return
			}
			if maxUpgrades < 1 {
				t.Skipf("no upgrades possible from %s to %s", minimumSupported, fv)
				return
			}

			for numUpgrades := 1; numUpgrades <= maxUpgrades; numUpgrades++ {
				t.Run(fmt.Sprintf("numUpgrades=%d", numUpgrades), func(t *testing.T) {

					// --- No skips: fully deterministic, single path ---
					t.Run("no-skip", func(t *testing.T) {
						mvt := newTest(
							NumUpgrades(numUpgrades),
							DisableSkipVersionUpgrades,
						)
						mvt.options.predecessorFunc = latestPredecessor
						assertValidTest(mvt, t.Fatal)

						path, err := mvt.chooseUpgradePath(numUpgrades, false)
						require.NoError(t, err)
						verifyPathProperties(t, path, numUpgrades, false,
							mvt.options.minimumBootstrapVersion,
							mvt.options.minimumSupportedVersion)
						totalPaths++
					})

					// --- With skips: enumerate all 2^n combinations ---
					t.Run("skip", func(t *testing.T) {
						// Determine how many predecessors the chain has by
						// running without skips first.
						mvt := newTest(
							NumUpgrades(numUpgrades),
							DisableSkipVersionUpgrades,
						)
						mvt.options.predecessorFunc = latestPredecessor
						assertValidTest(mvt, t.Fatal)

						basePath, err := mvt.chooseUpgradePath(numUpgrades, false)
						require.NoError(t, err)

						// predecessors = path minus the trailing current version
						n := len(basePath) - 1
						if n <= 1 {
							// chooseSkips returns early when <= 1 predecessor.
							t.Skipf("only %d predecessor(s), no skip combinations", n)
							return
						}

						uniquePaths := make(map[string]bool)

						// Enumerate all 2^n bit patterns.
						for mask := 0; mask < (1 << n); mask++ {
							mvt := newTest(
								NumUpgrades(numUpgrades),
								WithSkipVersionProbability(1),
							)
							mvt.options.predecessorFunc = latestPredecessor
							assertValidTest(mvt, t.Fatal)
							// Replace the PRNG with our controlled source so
							// that chooseSkips produces a specific skip pattern.
							mvt.prng = rand.New(
								&controlledSource{values: bitsToSourceValues(mask, n)},
							)

							path, err := mvt.chooseUpgradePath(numUpgrades, true)
							require.NoError(t, err)
							verifyPathProperties(t, path, numUpgrades, true,
								mvt.options.minimumBootstrapVersion,
								mvt.options.minimumSupportedVersion)

							key := pathToString(path)
							uniquePaths[key] = true
						}

						totalPaths += len(uniquePaths)
						t.Logf("%d unique paths from %d combinations (%d predecessors)",
							len(uniquePaths), 1<<n, n)

						// Sanity: with skip combinations, we should discover
						// at least 1 path (the no-skip base path) and at most
						// 2^n paths.
						require.GreaterOrEqual(t, len(uniquePaths), 1)
						require.LessOrEqual(t, len(uniquePaths), 1<<n)
					})
				})
			}
		})
	}

	t.Logf("total unique paths verified: %d", totalPaths)
}

// Test_chooseUpgradePath_exhaustive_with_msv extends the exhaustive
// test to exercise explicit minimumSupportedVersion constraints.
// The key property verified is that the MSV-series predecessor is
// never skipped.
func Test_chooseUpgradePath_exhaustive_with_msv(t *testing.T) {
	defer setDefaultVersions()()
	defer withTestBuildVersion("v25.2.0")()

	// MSV configurations: each is a (msv, mbv) pair where mbv <= msv.
	msvConfigs := []struct {
		msv string
		mbv string
	}{
		{msv: "v24.1.1", mbv: "v23.1.1"},
		{msv: "v24.3.1", mbv: "v24.1.1"},
		{msv: "v24.3.1", mbv: "v23.2.1"},
		{msv: "v24.1.1", mbv: "v24.1.1"},
	}

	for _, cfg := range msvConfigs {
		name := fmt.Sprintf("msv=%s/mbv=%s", cfg.msv, cfg.mbv)
		t.Run(name, func(t *testing.T) {
			currentV := clusterupgrade.CurrentVersion()
			mbv := clusterupgrade.MustParseVersion(cfg.mbv)
			maxUpgrades, err := release.MajorReleasesBetween(
				&currentV.Version, &mbv.Version)
			require.NoError(t, err)

			for numUpgrades := 1; numUpgrades <= maxUpgrades; numUpgrades++ {
				t.Run(fmt.Sprintf("numUpgrades=%d", numUpgrades), func(t *testing.T) {
					mvt := newTest(
						NumUpgrades(numUpgrades),
						MinimumSupportedVersion(cfg.msv),
						MinimumBootstrapVersion(cfg.mbv),
						DisableSkipVersionUpgrades,
					)
					mvt.options.predecessorFunc = latestPredecessor
					assertValidTest(mvt, t.Fatal)

					basePath, err := mvt.chooseUpgradePath(numUpgrades, false)
					require.NoError(t, err)
					verifyPathProperties(t, basePath, numUpgrades, false,
						mvt.options.minimumBootstrapVersion,
						mvt.options.minimumSupportedVersion)

					n := len(basePath) - 1
					if n <= 1 {
						return
					}

					// Enumerate all skip combinations.
					for mask := 0; mask < (1 << n); mask++ {
						mvt := newTest(
							NumUpgrades(numUpgrades),
							MinimumSupportedVersion(cfg.msv),
							MinimumBootstrapVersion(cfg.mbv),
							WithSkipVersionProbability(1),
						)
						mvt.options.predecessorFunc = latestPredecessor
						assertValidTest(mvt, t.Fatal)
						mvt.prng = rand.New(
							&controlledSource{values: bitsToSourceValues(mask, n)},
						)

						path, err := mvt.chooseUpgradePath(numUpgrades, true)
						require.NoError(t, err)
						verifyPathProperties(t, path, numUpgrades, true,
							mvt.options.minimumBootstrapVersion,
							mvt.options.minimumSupportedVersion)
					}
				})
			}
		})
	}
}
