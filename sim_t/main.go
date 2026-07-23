package main

const sy = `
package transport

type String string

func (s String) decode() {

}

func (s String) encode() {
	
}

struct UserComponent : TcpPacket
{
	lendian int i;
	bendian int a;
}

packet Message : TcpPacket {
	big uint16 length
	byte payload[length]
	big uint32 checksum
}

component User {
	string name
}

packet UserRequest : TcpPacket {
	big uint32 id
}

packet UserComponent : TcpPacket {
	little int32 i
	big    uint32 a
}
	
struct UserComponent : TcpPacket
{
	lendian int i;
	bendian int a;
}

packet Message {
    be u16 length(data)
    bytes data
    cstr name
}

package protocol

packet UserComponent : TcpPacket {
    le i32 id
    be u16 size = len(data)
    bytes data[size]
    cstr name
}
`

func main() {
	println(12)
}
