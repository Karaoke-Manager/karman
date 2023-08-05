package mediatype

// These constants define known media types.
// Using these constants is not required but can be more convenient than parsing strings every time.
var (
	Any = MediaType{"*", "*", nil, 1} // */*

	TextAny   = MediaType{"text", "*", nil, 1}     // text/*
	TextPlain = MediaType{"text", "plain", nil, 1} // text/plain

	ImageAny  = MediaType{"image", "*", nil, 1}    // image/*
	ImageJPEG = MediaType{"image", "jpeg", nil, 1} // image/jpeg
	ImagePNG  = MediaType{"image", "png", nil, 1}  // image/png
	ImageGIF  = MediaType{"image", "gif", nil, 1}  // image/gif

	AudioMPEG = MediaType{"audio", "mpeg", nil, 1}

	VideoMP4 = MediaType{"video", "mp4", nil, 1}

	ApplicationJSON        = MediaType{"application", "json", nil, 1}         // application/json
	ApplicationProblemJSON = MediaType{"application", "problem+json", nil, 1} // application/problem+json
	ApplicationXML         = MediaType{"application", "xml", nil, 1}          // application/xml
	ApplicationProblemXML  = MediaType{"application", "problem+xml", nil, 1}  // application/problem+xml
)
