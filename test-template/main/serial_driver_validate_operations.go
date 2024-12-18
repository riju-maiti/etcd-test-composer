package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/anishathalye/porcupine"
	combinations "github.com/mxschmitt/golang-combinations"
	"github.com/antithesishq/antithesis-sdk-go/assert"
)

type kvInput struct {
	op    uint8 // 0 => get, 1 => put
	key   string
	value string
}

type kvOutput struct {
	value string
}

// this model partitions history by key
var kvModel = porcupine.Model{
	Partition: func(history []porcupine.Operation) [][]porcupine.Operation {
		m := make(map[string][]porcupine.Operation)
		for _, v := range history {
			key := v.Input.(kvInput).key
			m[key] = append(m[key], v)
		}
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		ret := make([][]porcupine.Operation, 0, len(keys))
		for _, k := range keys {
			ret = append(ret, m[k])
		}
		return ret
	},
	PartitionEvent: func(history []porcupine.Event) [][]porcupine.Event {
		m := make(map[string][]porcupine.Event)
		match := make(map[int]string) // id -> key
		for _, v := range history {
			if v.Kind == porcupine.CallEvent {
				key := v.Value.(kvInput).key
				m[key] = append(m[key], v)
				match[v.Id] = key
			} else {
				key := match[v.Id]
				m[key] = append(m[key], v)
			}
		}
		var ret [][]porcupine.Event
		for _, v := range m {
			ret = append(ret, v)
		}
		return ret
	},
	Init: func() interface{} {
		// note: we are modeling a single key's value here;
		// we're partitioning by key, so this is okay
		return ""
	},
	Step: func(state, input, output interface{}) (bool, interface{}) {
		inp := input.(kvInput)
		out := output.(kvOutput)
		st := state.(string)
		if inp.op == 0 {
			// get
			return out.value == st, state
		} else if inp.op == 1 {
			// put
			return true, inp.value
		} else {
			// append
			return true, (st + inp.value)
		}
	},
	DescribeOperation: func(input, output interface{}) string {
		inp := input.(kvInput)
		out := output.(kvOutput)
		switch inp.op {
		case 0:
			return fmt.Sprintf("get('%s') -> '%s'", inp.key, out.value)
		case 1:
			return fmt.Sprintf("put('%s', '%s')", inp.key, inp.value)
		case 2:
			return fmt.Sprintf("append('%s', '%s')", inp.key, inp.value)
		default:
			return "<invalid>"
		}
	},
}

func Validate(filepath string, viz bool) {
	fmt.Println("antithesis-porcupine: Starting Validate()")
	assert.Sometimes(true, "antithesis-porcupine: Starting Validate()", nil)

	fmt.Println("antithesis-porcupine: Opening operations log file")
	file, err := os.Open(filepath)
	if err != nil {
		fmt.Printf("antithesis-porcupine: error reading etcd log, can't validate test \n")
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	operations := []porcupine.Operation{}
	failed_ops := []porcupine.Operation{}

	for scanner.Scan() {
		// {id},{op},{start},{end},{key},{value},{response},{success},{revision}
		// 0, 1, 2, 3, 4, 5, 6, 7, 8
		vals := strings.Split(scanner.Text(), ",")
		op := porcupine.Operation{}
		op.ClientId, err = strconv.Atoi(vals[0])
		if err != nil {
			op.ClientId = 0
		}
		op_key := vals[4]
		op_value := vals[5]
		op_response := vals[6]
		if op_value == "None" {
			// convert to empty string
			op_value = ""
		}
		if op_response == "None" {
			op_response = ""
		}
		// only support put and get for now, get is 0, put is 1
		if vals[1] == "put" {
			op.Input = kvInput{op: 1, key: op_key, value: op_value}
			op.Output = kvOutput{value: op_response}
		} else {
			// this should be a get
			op.Input = kvInput{op: 0, key: op_key}
			op.Output = kvOutput{value: op_response}
		}

		// now we are writing times as ints, so convert
		op.Call, err = strconv.ParseInt(vals[2], 10, 64)
		op.Return, err = strconv.ParseInt(vals[3], 10, 64)

		if vals[7] == "False" {
			// add to failed_ops to go through later
			fmt.Printf("Failed operation: ")
			failed_ops = append(failed_ops, op)
		} else {
			operations = append(operations, op)
		}

		fmt.Println(op)
	}

	// first test the successful operations to see if they are linearizable
	fmt.Printf("antithesis-porcupine: number of successful operations found: %v \n", len(operations))
	if len(operations) == 0 {
		fmt.Printf("antithesis-porcupine: no successful operations, can't validate\n")
		return
	}

	if !viz {
		result := porcupine.CheckOperations(kvModel, operations)
		if result == true {
			// the model is linearizable, no need to check failed ops
			fmt.Printf("antithesis-porcupine: Validate result %v \n", result)
			return
		}

		// now create all combinations of failed operations and go through them to check for linearizability
		// once we have found a case that works, return true
		// if no cases are linearizable, return false

		fmt.Printf("antithesis-porcupine: %v failed operations to check\n", len(failed_ops))

		failed_op_combos := combinations.All(failed_ops)
		fmt.Printf("antithesis-porcupine: %v total combinations to check\n", len(failed_op_combos))

		for _, ops := range failed_op_combos {
			fmt.Printf("adding %v failed ops in this round \n", len(ops))
			//fmt.Printf("antithesis-porcupine: failed ops in this round %v\n", ops)
			ops_to_check := append(operations, ops...)
			//fmt.Printf("length of ops to check: %v\n", len(ops_to_check))
			res := porcupine.CheckOperations(kvModel, ops_to_check)
			if res {
				fmt.Printf("antithesis-porcupine: Validate result true.\n")
				fmt.Printf("antithesis-porcupine: failed ops that were used for true result - %v.\n", ops)
				return
			}
		}

		// we made it through all possible combinations, so returning false
		fmt.Printf("antithesis-porcupine: Validate result false.\n")
		return
	} else {
		// this only creates a visualization for the successful operations
		// TODO add visualization for failed operations too?
		res, info := porcupine.CheckOperationsVerbose(kvModel, operations, 0)
		file, err := os.CreateTemp("resources", "*.html")
		if err != nil {
			fmt.Printf("failed to create temp file\n")
		}
		err = porcupine.Visualize(kvModel, info, file)
		if err != nil {
			fmt.Printf("visualization failed\n")
		}
		fmt.Printf("wrote visualization to %s\n", file.Name())

		assert.Always(res == porcupine.Ok, "Operations against the cluster are linearizable", nil)
	}
}

func main() {
	fmt.Println("antithesis-porcupine: entered validate/main()")

	op_log := "/opt/antithesis/local-txt-files/operations.txt"
	create_viz := false

	if len(os.Args) > 1 {
		op_log = os.Args[1]
		create_viz = true
	}

	Validate(op_log, create_viz)

	fmt.Println("antithesis-porcupine: Validate done")
}
