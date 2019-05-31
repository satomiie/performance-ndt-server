package model

// The Measurement struct contains measurement results. This structure is
// meant to be serialised as JSON as sent as a textual message.
type Measurement struct {
	// AppInfo contains application level measurements.
	AppInfo *AppInfo `json:"app_info,omitempty"`

	// ConnectionInfo contains connection information.
	ConnectionInfo *ConnectionInfo `json:"connection_info,omitempty"`

	// BBRInfo is the data measured using TCP BBR instrumentation.
	BBRInfo *BBRInfo `json:"bbr_info,omitempty"`

	// Elapsed is the number of seconds elapsed since the beginning.
	Elapsed float64 `json:"elapsed"`

	// TCPInfo contains metrics measured using TCP_INFO instrumentation.
	TCPInfo *TCPInfo `json:"tcp_info,omitempty"`
}
