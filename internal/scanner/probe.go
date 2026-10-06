package scanner

import "encoding/json"

type probeJSON struct {
	Format struct {
		Duration string `json:"duration"`
	} `json:"format"`
	Streams []struct {
		CodecType string `json:"codec_type"`
		CodecName string `json:"codec_name"`
		Width     int    `json:"width"`
		Height    int    `json:"height"`
	} `json:"streams"`
}

func parseProbe(data []byte) *mediaInfo {
	var pj probeJSON
	if err := json.Unmarshal(data, &pj); err != nil {
		return nil
	}
	mi := &mediaInfo{}
	for _, st := range pj.Streams {
		if st.CodecType == "video" && mi.VCodec == "" {
			mi.VCodec, mi.Width, mi.Height = st.CodecName, st.Width, st.Height
		}
		if st.CodecType == "audio" && mi.ACodec == "" {
			mi.ACodec = st.CodecName
		}
	}
	var d float64
	json.Unmarshal([]byte(`"`+pj.Format.Duration+`"`), &d)
	mi.Duration = d
	return mi
}
