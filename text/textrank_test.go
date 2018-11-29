package text_test

import (
	"fmt"
	"testing"

	"github.com/RTradeLtd/Lens/text"
)

const (
	testText1 = "Temporal is a primarily hosted API into distributed/decentralized storage technologies, with the focus on making it easy for the end user to integrate with the next generation of data storage protocols. While primarily hosted, we are pursuing an open-source model similar to Zabbix in which the code is open source, and we encourage people to run their own TEMPORAL instances. However, for people who choose to host their own instances, we will offer enterprise support, training, maintenance and implementation packages. For those that don't have the technical capacity to do so, they may use our paid for hosted API. While the hosted API service is centralized in the sense that users are relying on us to provide a server, and act as hosts of the data, we plug into the public networks of supported protocols by default, allowing users to leverage the power of the public networks, and ensuring that we aren't the only holders of the data."
	testText2 = "Paying for the hosted TEMPORAL service will be done on a per-interaction basis through the web interface orchestrated through smart contracts. Until the beta is finished, all TEMPORAL access will be free of charge. However, this comes with the condition that there is no guarantee in data persistence guarantees. Should you wish to use TEMPORAL but have data persistence, please contact us and we can set something up."
)

func TestTextAnalyzer(t *testing.T) {
	ta := text.NewTextAnalyzer(true)
	summary := ta.Summarize(testText1, 0.02)
	fmt.Println(summary)
	ta.Clear()
	summary = ta.Summarize(testText2, 0.02)
}
