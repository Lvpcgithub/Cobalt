package config

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
)

type Packet struct {
	Length      uint16   // 完整数据包长度
	HeaderLen   uint16   // 自定义包头信息长度
	Timestamp   uint32   // 时间戳
	PacketID    uint32   // 合并请求的唯一标识ID
	PacketType  uint8    // 请求类型
	Property    uint16   // 流的时延或带宽需求
	Priority    uint8    // 优先级
	HopCounts   uint8    // 当前在第几跳
	PacketCount uint8    // 合并的请求数量
	Offsets     []uint8  // 每个请求的偏移量
	Padding     []uint8  // 填充
	HopList     []uint32 // 完整转发路径 (每个 IP 地址以 uint32 表示)
}

// 将字符串形式的 IP 转换为 uint32
func ipToUint32(ip string) (uint32, error) {
	parsedIP := net.ParseIP(ip).To4()
	if parsedIP == nil {
		return 0, fmt.Errorf("无效的 IP 地址: %s", ip)
	}
	return binary.BigEndian.Uint32(parsedIP), nil
}

// 将 uint32 转换为字符串形式的 IP
func uint32ToIP(ipUint uint32) string {
	return fmt.Sprintf("%d.%d.%d.%d",
		byte(ipUint>>24),
		byte(ipUint>>16),
		byte(ipUint>>8),
		byte(ipUint),
	)
}
func SerializePacket(packet *Packet) ([]byte, error) {
	buffer := new(bytes.Buffer)

	// 按顺序写入固定大小的字段
	err := binary.Write(buffer, binary.BigEndian, packet.Length)
	if err != nil {
		return nil, err
	}

	err = binary.Write(buffer, binary.BigEndian, packet.HeaderLen)
	if err != nil {
		return nil, err
	}

	err = binary.Write(buffer, binary.BigEndian, packet.Timestamp)
	if err != nil {
		return nil, err
	}

	err = binary.Write(buffer, binary.BigEndian, packet.PacketID)
	if err != nil {
		return nil, err
	}

	err = binary.Write(buffer, binary.BigEndian, packet.PacketType)
	if err != nil {
		return nil, err
	}

	err = binary.Write(buffer, binary.BigEndian, packet.Property)
	if err != nil {
		return nil, err
	}

	err = binary.Write(buffer, binary.BigEndian, packet.Priority)
	if err != nil {
		return nil, err
	}

	err = binary.Write(buffer, binary.BigEndian, packet.HopCounts)
	if err != nil {
		return nil, err
	}

	err = binary.Write(buffer, binary.BigEndian, packet.PacketCount)
	if err != nil {
		return nil, err
	}

	// 写入 Offsets
	for _, offset := range packet.Offsets {
		err = binary.Write(buffer, binary.BigEndian, offset)
		if err != nil {
			return nil, err
		}
	}

	// 写入 Padding
	for _, pad := range packet.Padding {
		err = binary.Write(buffer, binary.BigEndian, pad)
		if err != nil {
			return nil, err
		}
	}

	// 写入 HopList (每个 IP 地址为 uint32)
	for _, hop := range packet.HopList {
		err = binary.Write(buffer, binary.BigEndian, hop)
		if err != nil {
			return nil, err
		}
	}

	return buffer.Bytes(), nil
}
func DeserializePacket(data []byte) (*Packet, error) {
	buffer := bytes.NewReader(data)

	packet := &Packet{}

	// 解析固定大小的字段
	err := binary.Read(buffer, binary.BigEndian, &packet.Length)
	if err != nil {
		return nil, err
	}

	err = binary.Read(buffer, binary.BigEndian, &packet.HeaderLen)
	if err != nil {
		return nil, err
	}

	err = binary.Read(buffer, binary.BigEndian, &packet.Timestamp)
	if err != nil {
		return nil, err
	}

	err = binary.Read(buffer, binary.BigEndian, &packet.PacketID)
	if err != nil {
		return nil, err
	}

	err = binary.Read(buffer, binary.BigEndian, &packet.PacketType)
	if err != nil {
		return nil, err
	}

	err = binary.Read(buffer, binary.BigEndian, &packet.Property)
	if err != nil {
		return nil, err
	}

	err = binary.Read(buffer, binary.BigEndian, &packet.Priority)
	if err != nil {
		return nil, err
	}

	err = binary.Read(buffer, binary.BigEndian, &packet.HopCounts)
	if err != nil {
		return nil, err
	}

	err = binary.Read(buffer, binary.BigEndian, &packet.PacketCount)
	if err != nil {
		return nil, err
	}

	// 打印调试信息
	fmt.Printf("反序列化 HeaderLen: %d, Length: %d, HopCounts: %d, PacketCount: %d\n",
		packet.HeaderLen, packet.Length, packet.HopCounts, packet.PacketCount)

	// 解析 Offsets
	packet.Offsets = make([]uint8, packet.PacketCount)
	for i := 0; i < int(packet.PacketCount); i++ {
		var offset uint8
		err = binary.Read(buffer, binary.BigEndian, &offset)
		if err != nil {
			return nil, err
		}
		packet.Offsets[i] = offset
	}

	// 计算 Padding 长度
	fixedFieldSize := 12 + int(packet.PacketCount) // 固定字段 + Offsets 长度
	paddingLength := int(packet.HeaderLen) - fixedFieldSize
	if paddingLength > 0 {
		packet.Padding = make([]uint8, paddingLength)
		err := binary.Read(buffer, binary.BigEndian, packet.Padding)
		if err != nil {
			fmt.Printf("读取 Padding 时出错: %v\n", err)
			return nil, err
		}
		fmt.Printf("反序列化 Padding: %v\n", packet.Padding)
	}

	// 解析 HopList
	packet.HopList = make([]uint32, packet.HopCounts)
	for i := 0; i < int(packet.HopCounts); i++ {
		var hop uint32
		err := binary.Read(buffer, binary.BigEndian, &hop)
		if err != nil {
			fmt.Printf("读取 HopList[%d] 时出错: %v\n", i, err)
			return nil, err
		}
		fmt.Printf("反序列化 HopList[%d]: %d\n", i, hop)
		packet.HopList[i] = hop
	}

	return packet, nil
}
