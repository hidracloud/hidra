package tcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"reflect"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hidracloud/hidra/v3/internal/metrics"
	"github.com/hidracloud/hidra/v3/internal/plugins"
)

type TcpPortsScenario struct {
	plugins.BasePlugin
}

// IsPortOpen checks if a port is open or not.
func IsPortOpen(protocol string, host string, port uint16) bool {
	conn, err := net.DialTimeout(protocol, host+":"+strconv.Itoa(int(port)), time.Second)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}

func UniquePorts(intSlices ...[]uint16) []uint16 {
	uniqueMap := map[uint16]bool{}

	for _, intSlice := range intSlices {
		for _, number := range intSlice {
			uniqueMap[number] = true
		}
	}

	// Create a slice with the capacity of unique items
	// This capacity make appending flow much more efficient
	result := make([]uint16, 0, len(uniqueMap))

	for key := range uniqueMap {
		result = append(result, key)
	}

	// sort expected ports in ascending order
	sort.Slice(result, func(i, j int) bool {
		return result[i] < result[j]
	})

	return result
}

func CheckOpenPorts(protocol string, host string, checkPorts chan uint16, taskToRun uint32, workers int) (openedPorts []uint16) {
	var wg sync.WaitGroup
	var mutex sync.Mutex

	defer close(checkPorts)

	ranTasks := uint32(0)
	for i := 0; i < workers; i++ {
		// increment worker count
		wg.Add(1)

		// worker starts here
		go func() {
			// we will notify this worker finished work
			defer wg.Done()

			// worker loop
			for {
				// increment processed tasks
				atomic.AddUint32(&ranTasks, 1)

				if ranTasks >= taskToRun {
					break
				}

				// port we are checking
				port := <-checkPorts

				// check it
				if IsPortOpen(protocol, host, port) {
					mutex.Lock()

					// it is open, add it to the list
					openedPorts = append(openedPorts, port)

					mutex.Unlock()
				}
			}
		}()
	}

	// wait for workers
	wg.Wait()

	// sort opened ports in ascending order
	sort.Slice(openedPorts, func(i, j int) bool {
		return openedPorts[i] < openedPorts[j]
	})

	return openedPorts
}

// CheckOpenPortsRange checks from a range of ports and returns the opened ports
func CheckOpenPortsRange(protocol string, host string, start, end uint16, workers int) []uint16 {
	// how much work we will do
	taskToRun := (end - start) + 1

	// make chan with all ports to check
	var checkPorts = make(chan uint16, taskToRun)
	for i := start; i <= end; i++ {
		checkPorts <- i
	}

	return CheckOpenPorts(protocol, host, checkPorts, uint32(taskToRun), workers)
}

// CheckOpenPorts checks from a slice of ports and returns the opened ports
func CheckOpenPortsFromSlice(protocol string, host string, ports []uint16, workers int) (openedPorts []uint16) {
	// how much work we will do
	taskToRun := len(ports)

	// make chan with all ports to check
	var checkPorts = make(chan uint16, taskToRun)
	for _, port := range ports {
		checkPorts <- port
	}

	return CheckOpenPorts(protocol, host, checkPorts, uint32(taskToRun), workers)
}

func getExpectedPorts(ports string) []uint16 {
	// split ports by comma
	tmp := strings.Split(ports, ",")
	values := make([]uint16, 0, len(tmp))

	// convert from string to uint16
	for _, raw := range tmp {
		v, err := strconv.Atoi(raw)
		if err != nil {
			log.Print(err)
			continue
		}
		values = append(values, uint16(v))
	}

	// sort expected ports in ascending order
	sort.Slice(values, func(i, j int) bool {
		return values[i] < values[j]
	})

	return values
}

func (s *TcpPortsScenario) checkOpenPorts(ctx2 context.Context, args map[string]string, stepsgen map[string]any) ([]*metrics.Metric, error) {
	// find opened ports from a 10k random range
	check_per_round := uint16(10000)

	// we will try to find unknow ports opened
	// we can't check all because scenario would be too slow
	from_port := uint16(uint16(rand.Intn(65000)-int(check_per_round)) + 1)
	to_port := uint16(from_port + check_per_round)

	// if to_port overflows
	if to_port < from_port {
		to_port = 0xffff - 1
	}

	// target max 30 seconds of execution
	workers := (int(to_port-from_port) / 30)

	// expected opened ports
	expectedPorts := getExpectedPorts(args["ports"])

	// check specific known ports
	openedPorts1 := CheckOpenPortsFromSlice("tcp4", args["host"], expectedPorts, runtime.NumCPU())

	// check a bunch of ports
	openedPorts2 := CheckOpenPortsRange("tcp4", args["host"], from_port, to_port, workers)

	// merge slices
	openedPorts := UniquePorts(openedPorts1, openedPorts2)

	// exported metrics
	customMetrics := []*metrics.Metric{
		{
			Name:        "tcp_open_ports_mismatch",
			Description: "If we found a mismatch value will be 1",
			Value:       0,
		},
	}

	// compare slices
	if reflect.DeepEqual(openedPorts, expectedPorts) {
		// equally the same :)
		return customMetrics, nil
	}

	// we found a mismatch
	customMetrics[0].Value = 1

	// trigger expected/received mismatch
	s1, _ := json.Marshal(openedPorts)
	s2, _ := json.Marshal(expectedPorts)
	return customMetrics, fmt.Errorf("data expected: %s, data received: %s", string(s2), string(s1))
}

// Init initialize the scenario
func (s *TcpPortsScenario) Init() {
	s.Primitives()

	s.RegisterStep(&plugins.StepDefinition{
		Name:        "opened",
		Description: "Check host opened ports",
		Params: []plugins.StepParam{
			{
				Name:        "host",
				Description: "Host to check ports on",
				Optional:    false,
			},
			{
				Name:        "ports",
				Description: "Ports expected to be open",
				Optional:    false,
			},
		},
		Fn: s.checkOpenPorts,
	})

}

// Init initializes the plugin.
func init() {
	h := &TcpPortsScenario{}
	h.Init()
	plugins.AddPlugin("tcp", "TCP plugin is used to check TCP opened ports at a server", h)
}
