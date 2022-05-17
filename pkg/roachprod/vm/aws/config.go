package aws

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"encoding/json"
	"io/ioutil"
	"sort"
	"strings"
)

type awsConfig struct {
	regions []awsRegion

	azByName map[string]*availabilityZone
}

type awsRegion struct {
	Name              string            `json:"region"`
	SecurityGroup     string            `json:"security_group"`
	AMI               string            `json:"ami_id"`
	AvailabilityZones availabilityZones `json:"subnets"`
}

type availabilityZone struct {
	name     string
	subnetID string
	region   *awsRegion
}

func (c *awsConfig) UnmarshalJSON(data []byte) error {
	__antithesis_instrumentation__.Notify(182855)
	type raw struct {
		Regions struct {
			Value []awsRegion `json:"value"`
		} `json:"regions"`
	}
	var v raw
	if err := json.Unmarshal(data, &v); err != nil {
		__antithesis_instrumentation__.Notify(182859)
		return err
	} else {
		__antithesis_instrumentation__.Notify(182860)
	}
	__antithesis_instrumentation__.Notify(182856)
	*c = awsConfig{
		regions:  v.Regions.Value,
		azByName: make(map[string]*availabilityZone),
	}
	sort.Slice(c.regions, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(182861)
		return c.regions[i].Name < c.regions[j].Name
	})
	__antithesis_instrumentation__.Notify(182857)
	for i := range c.regions {
		__antithesis_instrumentation__.Notify(182862)
		r := &c.regions[i]
		for i := range r.AvailabilityZones {
			__antithesis_instrumentation__.Notify(182863)
			az := &r.AvailabilityZones[i]
			az.region = r
			c.azByName[az.name] = az
		}
	}
	__antithesis_instrumentation__.Notify(182858)
	return nil
}

func (c *awsConfig) getRegion(name string) *awsRegion {
	__antithesis_instrumentation__.Notify(182864)
	i := sort.Search(len(c.regions), func(i int) bool {
		__antithesis_instrumentation__.Notify(182867)
		return c.regions[i].Name >= name
	})
	__antithesis_instrumentation__.Notify(182865)
	if i < len(c.regions) && func() bool {
		__antithesis_instrumentation__.Notify(182868)
		return c.regions[i].Name == name == true
	}() == true {
		__antithesis_instrumentation__.Notify(182869)
		return &c.regions[i]
	} else {
		__antithesis_instrumentation__.Notify(182870)
	}
	__antithesis_instrumentation__.Notify(182866)
	return nil
}

func (c *awsConfig) regionNames() (names []string) {
	__antithesis_instrumentation__.Notify(182871)
	for _, r := range c.regions {
		__antithesis_instrumentation__.Notify(182873)
		names = append(names, r.Name)
	}
	__antithesis_instrumentation__.Notify(182872)
	return names
}

func (c *awsConfig) getAvailabilityZone(azName string) *availabilityZone {
	__antithesis_instrumentation__.Notify(182874)
	return c.azByName[azName]
}

func (c *awsConfig) availabilityZoneNames() (zoneNames []string) {
	__antithesis_instrumentation__.Notify(182875)
	for _, r := range c.regions {
		__antithesis_instrumentation__.Notify(182877)
		for _, az := range r.AvailabilityZones {
			__antithesis_instrumentation__.Notify(182878)
			zoneNames = append(zoneNames, az.name)
		}
	}
	__antithesis_instrumentation__.Notify(182876)
	sort.Strings(zoneNames)
	return zoneNames
}

type availabilityZones []availabilityZone

func (s *availabilityZones) UnmarshalJSON(data []byte) error {
	__antithesis_instrumentation__.Notify(182879)
	var m map[string]string
	if err := json.Unmarshal(data, &m); err != nil {
		__antithesis_instrumentation__.Notify(182883)
		return err
	} else {
		__antithesis_instrumentation__.Notify(182884)
	}
	__antithesis_instrumentation__.Notify(182880)
	*s = make(availabilityZones, 0, len(m))
	for az, sn := range m {
		__antithesis_instrumentation__.Notify(182885)
		*s = append(*s, availabilityZone{
			name:     az,
			subnetID: sn,
		})
	}
	__antithesis_instrumentation__.Notify(182881)
	sort.Slice(*s, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(182886)
		return (*s)[i].name < (*s)[j].name
	})
	__antithesis_instrumentation__.Notify(182882)
	return nil
}

type awsConfigValue struct {
	path string
	awsConfig
}

func (c *awsConfigValue) Set(path string) (err error) {
	__antithesis_instrumentation__.Notify(182887)
	if path == "" {
		__antithesis_instrumentation__.Notify(182891)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(182892)
	}
	__antithesis_instrumentation__.Notify(182888)
	c.path = path
	var data []byte
	if strings.HasPrefix(path, "embedded:") {
		__antithesis_instrumentation__.Notify(182893)
		data, err = Asset(path[strings.Index(path, ":")+1:])
	} else {
		__antithesis_instrumentation__.Notify(182894)
		data, err = ioutil.ReadFile(path)
	}
	__antithesis_instrumentation__.Notify(182889)
	if err != nil {
		__antithesis_instrumentation__.Notify(182895)
		return err
	} else {
		__antithesis_instrumentation__.Notify(182896)
	}
	__antithesis_instrumentation__.Notify(182890)
	return json.Unmarshal(data, &c.awsConfig)
}

func (c *awsConfigValue) Type() string {
	__antithesis_instrumentation__.Notify(182897)
	return "aws config path"
}

func (c *awsConfigValue) String() string {
	__antithesis_instrumentation__.Notify(182898)
	if c.path == "" {
		__antithesis_instrumentation__.Notify(182900)
		return "see config.json"
	} else {
		__antithesis_instrumentation__.Notify(182901)
	}
	__antithesis_instrumentation__.Notify(182899)
	return c.path
}
