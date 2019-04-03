package common

import "time"

// Data struct
type Data struct {
	ID          string    `json:"id"`
	URL         string    `json:"url"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	PublishTime time.Time `json:"publish_time"`
	Author      string    `json:"author"`
	Source      string    `json:"source"`
	Tag         string    `json:"tag"`
	Lrclink     string    `json:"lrclink"`
	PicLink     string    `json:"pic_link"`
	FileLink    string    `json:"file_link"`
}
