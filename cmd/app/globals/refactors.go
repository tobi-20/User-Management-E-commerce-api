package globals

// import (
// 	"fmt"
// 	"net/http"
// )

// func SendEmail(w http.ResponseWriter, params EmailSendParams, sendFunc func(string, string) error) {
// 	link := fmt.Sprintf("%s%s?selector=%s&verifier=%s", SitePath, params.Path, params.Selector, params.Verifier)

// 	err := sendFunc(link, params.EmailAddr)
// 	if err != nil {
// 		http.Error(w, "link failed to generate", http.StatusInternalServerError)
// 		return
// 	}

// }
