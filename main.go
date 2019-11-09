package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	metric "github.com/phlipse/influxmetric"
	sap "github.com/phlipse/go-sapcontrol"
)

var (
	// global name for measurement name in metrics
	measurementName string
	// global boolean if to Replace spaces in tag keys/values and field keys with delimiter
	replaceSpaces bool
	// global delimeter
	delim string
	// global name for tag where node path or process list is stored, if set to ROOT the name of the root node will be used
	tagName string
	// global boolean if to be verbose
	verbose bool
	// global anRootNode is the number of the node where we want to start from with metric extraction
	anRootNode int
	// global boolean if to use timestamp SAP CCMS or not
	anSapTime bool
	// global string to decide for which field keys the SAP timestamp should be set
	anSapTimeFields string
)

var (
	// BUILD_VERSION is set during compile time
	BUILD_VERSION = "unknown"
	// BUILD_DATE is set during compile time
	BUILD_DATE = "unknown"
)

func init() {
	// general
	flag.StringVar(&measurementName, "m", "sapmon", "Metric measurement name.")
	flag.BoolVar(&replaceSpaces, "r", false, "Replace spaces in tag keys/values and field keys with delimiter, see also parameter \"d\". (default false)")
	flag.StringVar(&delim, "d", " ", "Delimiter to separate node names and replace spaces in tag keys/values and field keys, see also parameter \"r\".")
	flag.StringVar(&tagName, "p", "ccms", "Name of tag where node path or process list is stored. If set to ROOT the name of the root node will be used.")
	flag.BoolVar(&verbose, "v", false, "Verbose output. (default false)")
	// only AlertNodes
	flag.IntVar(&anRootNode, "n", 0, "Number of node which is our root to start from. (default 0)")
	flag.BoolVar(&anSapTime, "t", false, "Use timestamp from SAP AlertNode for metric. (default false)")
	flag.StringVar(&anSapTimeFields, "u", "all", "Field keys for which the SAP AlertNode timestamp should be set. Options: string, float, all = other input.")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Build version:\t%s\n", BUILD_VERSION)
		fmt.Fprintf(os.Stderr, "Build date:\t%s\n", BUILD_DATE)
		fmt.Fprintf(os.Stderr, "\nUsage:\n")
		flag.PrintDefaults()
	}
	flag.Parse()
}

func main() {
	// read in from stdin
	sapCtrl, err := sap.Read(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	// rSpaces replaces spaces with delimiter.
	rSpaces := strings.NewReplacer(" ", delim)

	var metrics metric.Metrics
	var pErr sap.ProcessingErrors

	switch sapCtrl.Function {
	case "GetProcessList":
		// get process list from sapcontrol output
		processList, err := sapCtrl.GetProcssList()
		pErr = err // could not assign directly to pErr because of variable scope

		// build up metrics
		for _, proc := range processList {
			// process name
			var m = metric.Metric{}
			m.TagSet = make(map[string]string)
			m.FieldSet = make(map[string]interface{})

			m.Measurement = measurementName

			// set process name as tag to identify metric
			tagKey := tagName
			tagValue := proc.Name
			if replaceSpaces {
				tagKey = rSpaces.Replace(tagKey)
				tagValue = rSpaces.Replace(tagValue)
			}
			m.TagSet[tagKey] = tagValue

			// set SAP status
			m.FieldSet["ActualValue"] = getStatusCode(proc.DispStatus)
			m.FieldSet["string"] = proc.TextStatus

			// set execution time of process
			m.FieldSet["float"] = proc.ElapsedTime.Seconds()

			// always take current time
			// process list has only a start time and an elapsed time which sometimes is far behind

			metrics = append(metrics, m)
		}
	case "GetAlertTree":
		// get nodes from sapcontrol output
		nodes, err := sapCtrl.GetAlertNodes()
		pErr = err // could not assign directly to pErr because of variable scope

		// get last nodes -> our metrics
		last := nodes.GetLastNodesByParentID(anRootNode)

		// build up metrics
		for _, node := range last {
			var m = metric.Metric{}
			m.TagSet = make(map[string]string)
			m.FieldSet = make(map[string]interface{})

			m.Measurement = measurementName

			// build up base part of tag value
			var tagKey, tagValue string
			switch tagName {
			case "ROOT":
				// handle special case ROOT
				tagKey = nodes.GetNodePath(node)[1].Name
				tagValue = nodes.GetNodePath(node)[2:].NodePathToName(delim)
			default:
				tagKey = tagName
				tagValue = nodes.GetNodePath(node)[1:].NodePathToName(delim)
			}

			// build up metric value
			v, u := sap.ParseValueUnit(node.Description)
			// check if we could parse a value/unit
			if u != node.Description {
				// build up unit, if we have one
				var unit string
				if u != "" {
					unit += fmt.Sprintf(" in %s", u)
				}
				// add unit to metric tag value
				tagValue += unit

				// we have to use floats because we could have a type mismatch in influxdb when having a GRAY status (string) followed by a GREEN (int/float)
				m.FieldSet["float"] = metric.MustFloat(v)
			} else {
				m.FieldSet["string"] = fmt.Sprintf("%v", metric.ExtractValue(node.Description))
			}

			// replace spaces in metric tag, if wanted
			if replaceSpaces {
				tagKey = rSpaces.Replace(tagKey)
				tagValue = rSpaces.Replace(tagValue)
			}

			// set metric tag
			m.TagSet[tagKey] = tagValue

			// set SAP status for node
			m.FieldSet["ActualValue"] = getStatusCode(node.ActualValue)

			// set timestamp, if wanted
			if anSapTime {
				// check if we have a string or float metric
				_, haveString := m.FieldSet["string"]
				_, haveFloat := m.FieldSet["float"]

				if (anSapTimeFields != "string" && anSapTimeFields != "float") ||
					(anSapTimeFields == "string" && haveString) ||
					(anSapTimeFields == "float" && haveFloat) {
					m.UnixTime = node.Time.UnixNano()
				}
			}

			metrics = append(metrics, m)
		}
	default:
		log.Fatalf("sapcontrol function \"%s\" not implemented yet", sapCtrl.Function)
	}

	// because we don't do any preprocessing of metrics, we could print it out directly
	for _, m := range metrics {
		fmt.Println(m)
	}

	// print out processing errors
	if verbose && len(pErr) > 0 {
		fmt.Printf("\n\nPROCESSING ERRORS:\n")
		for _, e := range pErr {
			fmt.Printf("%s\n", e.Message)
		}
	}
}
