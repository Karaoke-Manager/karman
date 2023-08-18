package uploads

/*
func (c *Controller) handleFileError(w http.ResponseWriter, r *http.Request, upload *model.Upload, path string, err error) {
	var details *apierror.ProblemDetails
	switch {
	case errors.Is(err, uploadSvc.ErrUploadClosed):
		details = apierror.UploadClosed(upload)
	case errors.Is(err, fs.ErrNotExist):
		details = apierror.UploadFileNotFound(upload, path)
	default:
		details = apierror.ErrInternalServerError
	}
	_ = render.Render(w, r, details)
}

func (c *Controller) PutFile(w http.ResponseWriter, r *http.Request) {
	upload := MustGetUpload(r.Context())
	path := MustGetFilePath(r.Context())

	if err := c.Service.CreateFile(r.Context(), upload, path, r.Body); err != nil {
		c.handleFileError(w, r, upload, path, err)
		return
	}

	_ = render.NoContent(w, r)
}

func (c *Controller) GetFile(w http.ResponseWriter, r *http.Request) {
	type FileSchema struct {
		render.NopRenderer
		Name  string `json:"name"`
		Size  int64  `json:"size"`
		IsDir bool   `json:"directory"`
	}

	type ResponseSchema struct {
		render.NopRenderer
		UploadUUID uuid.UUID    `json:"upload"`
		File       FileSchema   `json:"file"`
		Children   []FileSchema `json:"children,omitempty"`
	}

	upload := MustGetUpload(r.Context())
	path := MustGetFilePath(r.Context())

	stat, err := c.Service.StatFile(r.Context(), upload, path)
	if err != nil {
		c.handleFileError(w, r, upload, path, err)
		return
	}
	resp := ResponseSchema{
		UploadUUID: upload.UUID,
		File: FileSchema{
			Name:  stat.Name(),
			Size:  stat.Size(),
			IsDir: stat.IsDir(),
		},
	}
	if stat.IsDir() {
		entries, err := c.Service.ReadDir(r.Context(), upload, path+"/"+stat.Name())
		if err != nil {
			_ = render.Render(w, r, apierror.ErrInternalServerError)
			return
		}
		children := make([]FileSchema, len(entries))
		for i, entry := range entries {
			info, err := entry.Info()
			if err != nil {
				_ = render.Render(w, r, apierror.ErrInternalServerError)
				return
			}
			children[i] = FileSchema{
				Name:  entry.Name(),
				Size:  info.Size(),
				IsDir: entry.IsDir(),
			}
		}
		resp.Children = children
	}
	_ = render.Render(w, r, resp)
}

func (c *Controller) DeleteFile(w http.ResponseWriter, r *http.Request) {
	upload := MustGetUpload(r.Context())
	path := MustGetFilePath(r.Context())

	if err := c.Service.DeleteFile(r.Context(), upload, path); err != nil {
		c.handleFileError(w, r, upload, path, err)
		return
	}
	_ = render.NoContent(w, r)
}
*/
