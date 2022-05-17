package cluster

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"net"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/util/contextutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/log/severity"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/errors/oserror"
	"github.com/docker/distribution/reference"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/go-connections/nat"
	isatty "github.com/mattn/go-isatty"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
)

func dockerIP() net.IP {
	__antithesis_instrumentation__.Notify(20)
	host := os.Getenv("DOCKER_HOST")
	if host == "" {
		__antithesis_instrumentation__.Notify(25)
		host = client.DefaultDockerHost
	} else {
		__antithesis_instrumentation__.Notify(26)
	}
	__antithesis_instrumentation__.Notify(21)
	u, err := url.Parse(host)
	if err != nil {
		__antithesis_instrumentation__.Notify(27)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(28)
	}
	__antithesis_instrumentation__.Notify(22)
	if u.Scheme == "unix" {
		__antithesis_instrumentation__.Notify(29)
		return net.IPv4(127, 0, 0, 1)
	} else {
		__antithesis_instrumentation__.Notify(30)
	}
	__antithesis_instrumentation__.Notify(23)
	h, _, err := net.SplitHostPort(u.Host)
	if err != nil {
		__antithesis_instrumentation__.Notify(31)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(32)
	}
	__antithesis_instrumentation__.Notify(24)
	return net.ParseIP(h)
}

type Container struct {
	id      string
	name    string
	cluster *DockerCluster
}

func (c Container) Name() string {
	__antithesis_instrumentation__.Notify(33)
	return c.name
}

func hasImage(ctx context.Context, l *DockerCluster, ref string) error {
	__antithesis_instrumentation__.Notify(34)
	distributionRef, err := reference.ParseNamed(ref)
	if err != nil {
		__antithesis_instrumentation__.Notify(40)
		return err
	} else {
		__antithesis_instrumentation__.Notify(41)
	}
	__antithesis_instrumentation__.Notify(35)
	path := reference.Path(distributionRef)

	path = strings.TrimPrefix(path, "library/")

	images, err := l.client.ImageList(ctx, types.ImageListOptions{
		All: true,
		Filters: filters.NewArgs(
			filters.Arg("reference", path),
		),
	})
	if err != nil {
		__antithesis_instrumentation__.Notify(42)
		return err
	} else {
		__antithesis_instrumentation__.Notify(43)
	}
	__antithesis_instrumentation__.Notify(36)

	tagged, ok := distributionRef.(reference.Tagged)
	if !ok {
		__antithesis_instrumentation__.Notify(44)
		return errors.Errorf("untagged reference %s not permitted", ref)
	} else {
		__antithesis_instrumentation__.Notify(45)
	}
	__antithesis_instrumentation__.Notify(37)

	wanted := fmt.Sprintf("%s:%s", path, tagged.Tag())
	for _, image := range images {
		__antithesis_instrumentation__.Notify(46)
		for _, repoTag := range image.RepoTags {
			__antithesis_instrumentation__.Notify(47)

			if repoTag == wanted {
				__antithesis_instrumentation__.Notify(48)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(49)
			}
		}
	}
	__antithesis_instrumentation__.Notify(38)
	var imageList []string
	for _, image := range images {
		__antithesis_instrumentation__.Notify(50)
		for _, tag := range image.RepoTags {
			__antithesis_instrumentation__.Notify(51)
			imageList = append(imageList, "%s %s", tag, image.ID)
		}
	}
	__antithesis_instrumentation__.Notify(39)
	return errors.Errorf("%s not found in:\n%s", wanted, strings.Join(imageList, "\n"))
}

func pullImage(
	ctx context.Context, l *DockerCluster, ref string, options types.ImagePullOptions,
) error {
	__antithesis_instrumentation__.Notify(52)

	if hasImage(ctx, l, ref) == nil {
		__antithesis_instrumentation__.Notify(57)
		log.Infof(ctx, "ImagePull %s already exists", ref)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(58)
	}
	__antithesis_instrumentation__.Notify(53)

	log.Infof(ctx, "ImagePull %s starting", ref)
	defer log.Infof(ctx, "ImagePull %s complete", ref)

	rc, err := l.client.ImagePull(ctx, ref, options)
	if err != nil {
		__antithesis_instrumentation__.Notify(59)
		return err
	} else {
		__antithesis_instrumentation__.Notify(60)
	}
	__antithesis_instrumentation__.Notify(54)
	defer rc.Close()
	out := os.Stderr
	outFd := out.Fd()
	isTerminal := isatty.IsTerminal(outFd)

	if err := jsonmessage.DisplayJSONMessagesStream(rc, out, outFd, isTerminal, nil); err != nil {
		__antithesis_instrumentation__.Notify(61)
		return err
	} else {
		__antithesis_instrumentation__.Notify(62)
	}
	__antithesis_instrumentation__.Notify(55)
	if err := hasImage(ctx, l, ref); err != nil {
		__antithesis_instrumentation__.Notify(63)
		return errors.Wrapf(err, "pulled image %s but still don't have it", ref)
	} else {
		__antithesis_instrumentation__.Notify(64)
	}
	__antithesis_instrumentation__.Notify(56)
	return nil
}

func splitBindSpec(bind string) (hostPath string, containerPath string) {
	__antithesis_instrumentation__.Notify(65)
	s := strings.SplitN(bind, ":", 2)
	return s[0], s[1]
}

func getNonRootContainerUser() (string, error) {
	__antithesis_instrumentation__.Notify(66)

	const minUnreservedID = 101
	user, err := user.Current()
	if err != nil {
		__antithesis_instrumentation__.Notify(72)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(73)
	}
	__antithesis_instrumentation__.Notify(67)
	uid, err := strconv.Atoi(user.Uid)
	if err != nil {
		__antithesis_instrumentation__.Notify(74)
		return "", errors.Wrap(err, "looking up host UID")
	} else {
		__antithesis_instrumentation__.Notify(75)
	}
	__antithesis_instrumentation__.Notify(68)
	if uid < minUnreservedID {
		__antithesis_instrumentation__.Notify(76)
		return "", fmt.Errorf("host UID %d in container's reserved UID space", uid)
	} else {
		__antithesis_instrumentation__.Notify(77)
	}
	__antithesis_instrumentation__.Notify(69)
	gid, err := strconv.Atoi(user.Gid)
	if err != nil {
		__antithesis_instrumentation__.Notify(78)
		return "", errors.Wrap(err, "looking up host GID")
	} else {
		__antithesis_instrumentation__.Notify(79)
	}
	__antithesis_instrumentation__.Notify(70)
	if gid < minUnreservedID {
		__antithesis_instrumentation__.Notify(80)

		gid = uid
	} else {
		__antithesis_instrumentation__.Notify(81)
	}
	__antithesis_instrumentation__.Notify(71)
	return fmt.Sprintf("%d:%d", uid, gid), nil
}

func createContainer(
	ctx context.Context,
	l *DockerCluster,
	containerConfig container.Config,
	hostConfig container.HostConfig,
	platformSpec specs.Platform,
	containerName string,
) (*Container, error) {
	__antithesis_instrumentation__.Notify(82)
	hostConfig.NetworkMode = container.NetworkMode(l.networkID)

	hostConfig.DNSSearch = []string{"."}

	user, err := getNonRootContainerUser()
	if err != nil {
		__antithesis_instrumentation__.Notify(86)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(87)
	}
	__antithesis_instrumentation__.Notify(83)
	containerConfig.User = user

	for _, bind := range hostConfig.Binds {
		__antithesis_instrumentation__.Notify(88)
		hostPath, _ := splitBindSpec(bind)
		if _, err := os.Stat(hostPath); oserror.IsNotExist(err) {
			__antithesis_instrumentation__.Notify(89)
			maybePanic(os.MkdirAll(hostPath, 0755))
		} else {
			__antithesis_instrumentation__.Notify(90)
			maybePanic(err)
		}
	}
	__antithesis_instrumentation__.Notify(84)

	resp, err := l.client.ContainerCreate(ctx, &containerConfig, &hostConfig, nil, &platformSpec, containerName)
	if err != nil {
		__antithesis_instrumentation__.Notify(91)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(92)
	}
	__antithesis_instrumentation__.Notify(85)
	return &Container{
		id:      resp.ID,
		name:    containerName,
		cluster: l,
	}, nil
}

func maybePanic(err error) {
	__antithesis_instrumentation__.Notify(93)
	if err != nil {
		__antithesis_instrumentation__.Notify(94)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(95)
	}
}

func (c *Container) Remove(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(96)
	return c.cluster.client.ContainerRemove(ctx, c.id, types.ContainerRemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	})
}

func (c *Container) Kill(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(97)
	if err := c.cluster.client.ContainerKill(ctx, c.id, "9"); err != nil && func() bool {
		__antithesis_instrumentation__.Notify(99)
		return !strings.Contains(err.Error(), "is not running") == true
	}() == true {
		__antithesis_instrumentation__.Notify(100)
		return err
	} else {
		__antithesis_instrumentation__.Notify(101)
	}
	__antithesis_instrumentation__.Notify(98)
	c.cluster.expectEvent(c, eventDie)
	return nil
}

func (c *Container) Start(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(102)
	return c.cluster.client.ContainerStart(ctx, c.id, types.ContainerStartOptions{})
}

func (c *Container) Restart(ctx context.Context, timeout *time.Duration) error {
	__antithesis_instrumentation__.Notify(103)
	var exp []string
	if ci, err := c.Inspect(ctx); err != nil {
		__antithesis_instrumentation__.Notify(106)
		return err
	} else {
		__antithesis_instrumentation__.Notify(107)
		if ci.State.Running {
			__antithesis_instrumentation__.Notify(108)
			exp = append(exp, eventDie)
		} else {
			__antithesis_instrumentation__.Notify(109)
		}
	}
	__antithesis_instrumentation__.Notify(104)
	if err := c.cluster.client.ContainerRestart(ctx, c.id, timeout); err != nil {
		__antithesis_instrumentation__.Notify(110)
		return err
	} else {
		__antithesis_instrumentation__.Notify(111)
	}
	__antithesis_instrumentation__.Notify(105)
	c.cluster.expectEvent(c, append(exp, eventRestart)...)
	return nil
}

func (c *Container) Wait(ctx context.Context, condition container.WaitCondition) error {
	__antithesis_instrumentation__.Notify(112)
	waitOKBodyCh, errCh := c.cluster.client.ContainerWait(ctx, c.id, condition)
	select {
	case err := <-errCh:
		__antithesis_instrumentation__.Notify(113)
		return err
	case waitOKBody := <-waitOKBodyCh:
		__antithesis_instrumentation__.Notify(114)
		outputLog := filepath.Join(c.cluster.volumesDir, "logs", "console-output.log")
		cmdLog, err := os.Create(outputLog)
		if err != nil {
			__antithesis_instrumentation__.Notify(118)
			return err
		} else {
			__antithesis_instrumentation__.Notify(119)
		}
		__antithesis_instrumentation__.Notify(115)
		defer cmdLog.Close()

		out := io.MultiWriter(cmdLog, os.Stderr)
		if err := c.Logs(ctx, out); err != nil {
			__antithesis_instrumentation__.Notify(120)
			log.Warningf(ctx, "%v", err)
		} else {
			__antithesis_instrumentation__.Notify(121)
		}
		__antithesis_instrumentation__.Notify(116)

		if exitCode := waitOKBody.StatusCode; exitCode != 0 {
			__antithesis_instrumentation__.Notify(122)
			err = errors.Errorf("non-zero exit code: %d", exitCode)
			fmt.Fprintln(out, err.Error())
			log.Shoutf(ctx, severity.INFO, "command left-over files in %s", c.cluster.volumesDir)
		} else {
			__antithesis_instrumentation__.Notify(123)
		}
		__antithesis_instrumentation__.Notify(117)

		return err
	}
}

func (c *Container) Logs(ctx context.Context, w io.Writer) error {
	__antithesis_instrumentation__.Notify(124)
	rc, err := c.cluster.client.ContainerLogs(ctx, c.id, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
	})
	if err != nil {
		__antithesis_instrumentation__.Notify(127)
		return err
	} else {
		__antithesis_instrumentation__.Notify(128)
	}
	__antithesis_instrumentation__.Notify(125)
	defer rc.Close()

	for {
		__antithesis_instrumentation__.Notify(129)
		var header uint64
		if err := binary.Read(rc, binary.BigEndian, &header); err == io.EOF {
			__antithesis_instrumentation__.Notify(131)
			break
		} else {
			__antithesis_instrumentation__.Notify(132)
			if err != nil {
				__antithesis_instrumentation__.Notify(133)
				return err
			} else {
				__antithesis_instrumentation__.Notify(134)
			}
		}
		__antithesis_instrumentation__.Notify(130)
		size := header & math.MaxUint32
		if _, err := io.CopyN(w, rc, int64(size)); err != nil {
			__antithesis_instrumentation__.Notify(135)
			return err
		} else {
			__antithesis_instrumentation__.Notify(136)
		}
	}
	__antithesis_instrumentation__.Notify(126)
	return nil
}

func (c *Container) Inspect(ctx context.Context) (types.ContainerJSON, error) {
	__antithesis_instrumentation__.Notify(137)
	return c.cluster.client.ContainerInspect(ctx, c.id)
}

func (c *Container) Addr(ctx context.Context, port nat.Port) *net.TCPAddr {
	__antithesis_instrumentation__.Notify(138)
	containerInfo, err := c.Inspect(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(142)
		log.Errorf(ctx, "%v", err)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(143)
	}
	__antithesis_instrumentation__.Notify(139)
	bindings, ok := containerInfo.NetworkSettings.Ports[port]
	if !ok || func() bool {
		__antithesis_instrumentation__.Notify(144)
		return len(bindings) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(145)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(146)
	}
	__antithesis_instrumentation__.Notify(140)
	portNum, err := strconv.Atoi(bindings[0].HostPort)
	if err != nil {
		__antithesis_instrumentation__.Notify(147)
		log.Errorf(ctx, "%v", err)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(148)
	}
	__antithesis_instrumentation__.Notify(141)
	return &net.TCPAddr{
		IP:   dockerIP(),
		Port: portNum,
	}
}

type resilientDockerClient struct {
	client.APIClient
}

func (cli resilientDockerClient) ContainerStart(
	clientCtx context.Context, id string, opts types.ContainerStartOptions,
) error {
	__antithesis_instrumentation__.Notify(149)
	for {
		__antithesis_instrumentation__.Notify(150)
		err := contextutil.RunWithTimeout(clientCtx, "start container", 20*time.Second, func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(153)
			return cli.APIClient.ContainerStart(ctx, id, opts)
		})
		__antithesis_instrumentation__.Notify(151)

		if errors.Is(err, context.DeadlineExceeded) && func() bool {
			__antithesis_instrumentation__.Notify(154)
			return clientCtx.Err() == nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(155)
			log.Warningf(clientCtx, "ContainerStart timed out, retrying")
			continue
		} else {
			__antithesis_instrumentation__.Notify(156)
		}
		__antithesis_instrumentation__.Notify(152)
		return err
	}
}

func (cli resilientDockerClient) ContainerCreate(
	ctx context.Context,
	config *container.Config,
	hostConfig *container.HostConfig,
	networkingConfig *network.NetworkingConfig,
	platformSpec *specs.Platform,
	containerName string,
) (container.ContainerCreateCreatedBody, error) {
	__antithesis_instrumentation__.Notify(157)
	response, err := cli.APIClient.ContainerCreate(
		ctx, config, hostConfig, networkingConfig, platformSpec, containerName,
	)
	if err != nil && func() bool {
		__antithesis_instrumentation__.Notify(159)
		return strings.Contains(err.Error(), "already in use") == true
	}() == true {
		__antithesis_instrumentation__.Notify(160)
		log.Infof(ctx, "unable to create container %s: %v", containerName, err)
		containers, cerr := cli.ContainerList(ctx, types.ContainerListOptions{
			All:   true,
			Limit: -1,
		})
		if cerr != nil {
			__antithesis_instrumentation__.Notify(163)
			log.Infof(ctx, "unable to list containers: %v", cerr)
			return container.ContainerCreateCreatedBody{}, err
		} else {
			__antithesis_instrumentation__.Notify(164)
		}
		__antithesis_instrumentation__.Notify(161)
		for _, c := range containers {
			__antithesis_instrumentation__.Notify(165)
			for _, n := range c.Names {
				__antithesis_instrumentation__.Notify(166)

				n = strings.TrimPrefix(n, "/")
				if n != containerName {
					__antithesis_instrumentation__.Notify(169)
					continue
				} else {
					__antithesis_instrumentation__.Notify(170)
				}
				__antithesis_instrumentation__.Notify(167)
				log.Infof(ctx, "trying to remove %s", c.ID)
				options := types.ContainerRemoveOptions{
					RemoveVolumes: true,
					Force:         true,
				}
				if rerr := cli.ContainerRemove(ctx, c.ID, options); rerr != nil {
					__antithesis_instrumentation__.Notify(171)
					log.Infof(ctx, "unable to remove container: %v", rerr)
					return container.ContainerCreateCreatedBody{}, err
				} else {
					__antithesis_instrumentation__.Notify(172)
				}
				__antithesis_instrumentation__.Notify(168)
				return cli.ContainerCreate(ctx, config, hostConfig, networkingConfig, platformSpec, containerName)
			}
		}
		__antithesis_instrumentation__.Notify(162)
		log.Warningf(ctx, "error indicated existing container %s, "+
			"but none found:\nerror: %s\ncontainers: %+v",
			containerName, err, containers)

		return response, context.DeadlineExceeded
	} else {
		__antithesis_instrumentation__.Notify(173)
	}
	__antithesis_instrumentation__.Notify(158)
	return response, err
}
