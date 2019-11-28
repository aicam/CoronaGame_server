package routHandlers

//import (
//	"github.com/jinzhu/gorm"
//	"net"
//	"strings"
//)
//
//func JoinChatRoom(buf string, con net.Conn, db gorm.DB) int {
//	dispatched := strings.Split(buf, "/")
//	if len(dispatched) < 3 {
//		con.Write([]byte("error"))
//		return 0
//	}
//	user1 := dispatched[1]
//	user2 := dispatched[2]
//	return 0
//}
