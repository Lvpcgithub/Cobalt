package config

import (
	"bytes"
	"testing"
)

// 测试序列化和反序列化的功能
func TestSerializeAndDeserializePacket(t *testing.T) {
	// 创建一个示例 Packet 对象
	originalPacket := &Packet{
		Length:      64,
		HeaderLen:   20,
		Timestamp:   1672531200,
		PacketID:    12345678,
		PacketType:  1,
		Property:    256,
		Priority:    5,
		HopCounts:   2,
		PacketCount: 1,
		Offsets:     []uint8{10},
		Padding:     []uint8{0, 0, 0, 0},
		HopList:     []uint32{3232235777, 3232235778}, // 对应 IP: 192.168.1.1 和 192.168.1.2
	}

	// 测试序列化
	serializedData, err := SerializePacket(originalPacket)
	if err != nil {
		t.Fatalf("序列化失败: %v", err)
	}

	// 测试反序列化
	deserializedPacket, err := DeserializePacket(serializedData)
	if err != nil {
		t.Fatalf("反序列化失败: %v", err)
	}

	// 验证固定字段的值是否一致
	if originalPacket.Length != deserializedPacket.Length {
		t.Errorf("Length 不匹配: 原始值=%d, 反序列化值=%d", originalPacket.Length, deserializedPacket.Length)
	}
	if originalPacket.HeaderLen != deserializedPacket.HeaderLen {
		t.Errorf("HeaderLen 不匹配: 原始值=%d, 反序列化值=%d", originalPacket.HeaderLen, deserializedPacket.HeaderLen)
	}
	if originalPacket.Timestamp != deserializedPacket.Timestamp {
		t.Errorf("Timestamp 不匹配: 原始值=%d, 反序列化值=%d", originalPacket.Timestamp, deserializedPacket.Timestamp)
	}
	if originalPacket.PacketID != deserializedPacket.PacketID {
		t.Errorf("PacketID 不匹配: 原始值=%d, 反序列化值=%d", originalPacket.PacketID, deserializedPacket.PacketID)
	}
	if originalPacket.PacketType != deserializedPacket.PacketType {
		t.Errorf("PacketType 不匹配: 原始值=%d, 反序列化值=%d", originalPacket.PacketType, deserializedPacket.PacketType)
	}
	if originalPacket.Property != deserializedPacket.Property {
		t.Errorf("Property 不匹配: 原始值=%d, 反序列化值=%d", originalPacket.Property, deserializedPacket.Property)
	}
	if originalPacket.Priority != deserializedPacket.Priority {
		t.Errorf("Priority 不匹配: 原始值=%d, 反序列化值=%d", originalPacket.Priority, deserializedPacket.Priority)
	}
	if originalPacket.HopCounts != deserializedPacket.HopCounts {
		t.Errorf("HopCounts 不匹配: 原始值=%d, 反序列化值=%d", originalPacket.HopCounts, deserializedPacket.HopCounts)
	}
	if originalPacket.PacketCount != deserializedPacket.PacketCount {
		t.Errorf("PacketCount 不匹配: 原始值=%d, 反序列化值=%d", originalPacket.PacketCount, deserializedPacket.PacketCount)
	}

	// 验证切片字段是否一致
	if !bytes.Equal(originalPacket.Offsets, deserializedPacket.Offsets) {
		t.Errorf("Offsets 不匹配: 原始值=%v, 反序列化值=%v", originalPacket.Offsets, deserializedPacket.Offsets)
	}
	if !bytes.Equal(originalPacket.Padding, deserializedPacket.Padding) {
		t.Errorf("Padding 不匹配: 原始值=%v, 反序列化值=%v", originalPacket.Padding, deserializedPacket.Padding)
	}

	// 验证 HopList 是否一致
	if len(originalPacket.HopList) != len(deserializedPacket.HopList) {
		t.Errorf("HopList 长度不匹配: 原始长度=%d, 反序列化长度=%d", len(originalPacket.HopList), len(deserializedPacket.HopList))
	}
	for i, hop := range originalPacket.HopList {
		if hop != deserializedPacket.HopList[i] {
			t.Errorf("HopList[%d] 不匹配: 原始值=%d, 反序列化值=%d", i, hop, deserializedPacket.HopList[i])
		}
	}
}
