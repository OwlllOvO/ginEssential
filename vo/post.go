package vo

type CreatePostRequest struct {
	CategoryName string `json:"category_name" binding:"required"`
	Title        string `json:"title" binding:"required,max=10"`
	HeadImg      string `json:"head_img"`
	Content      string `json:"content" binding:"required"`
}
