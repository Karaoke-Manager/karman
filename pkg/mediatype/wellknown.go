package mediatype

// These constants define known media types.
// Using these constants is not required but can be more convenient than parsing strings every time.
var (
	Any = MediaType{"*", "*", nil} // */*

	TextAny   = MediaType{"text", "*", nil}     // text/*
	TextPlain = MediaType{"text", "plain", nil} // text/plain

	ImageAny  = MediaType{"image", "*", nil}    // image/*
	ImageJPEG = MediaType{"image", "jpeg", nil} // image/jpeg
	ImagePNG  = MediaType{"image", "png", nil}  // image/png
	ImageGIF  = MediaType{"image", "gif", nil}  // image/gif

	AudioMPEG = MediaType{"audio", "mpeg", nil}

	VideoMP4 = MediaType{"video", "mp4", nil}

	ApplicationJSON        = MediaType{"application", "json", nil}         // application/json
	ApplicationProblemJSON = MediaType{"application", "problem+json", nil} // application/problem+json
	ApplicationXML         = MediaType{"application", "xml", nil}          // application/xml
	ApplicationProblemXML  = MediaType{"application", "problem+xml", nil}  // application/problem+xml
)
