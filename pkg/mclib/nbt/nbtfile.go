package nbt

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
)

const (
	NodeTypeEnd       NodeType = 0
	NodeTypeByte      NodeType = 1
	NodeTypeShort     NodeType = 2
	NodeTypeInt       NodeType = 3
	NodeTypeLong      NodeType = 4
	NodeTypeFloat     NodeType = 5
	NodeTypeDouble    NodeType = 6
	NodeTypeByteArray NodeType = 7
	NodeTypeString    NodeType = 8
	NodeTypeList      NodeType = 9
	NodeTypeCompound  NodeType = 10
	NodeTypeIntArray  NodeType = 11
	NodeTypeLongArray NodeType = 12
)

type NodeType byte

type File struct {
	Root Node
}

type Node interface {
	Type() NodeType
}

func ReadFromFile(file string) (*File, error) {
	rawData, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	return ReadGZipFromStream(bytes.NewReader(rawData))
}

func ReadGZipFromStream(r io.Reader) (*File, error) {
	gzipReader, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("open gzip reader: %w", err)
	}

	return ReadFromStream(gzipReader)
}

func ReadFromStream(r io.Reader) (*File, error) {
	rootNode, err := readNodeOfType(r, NodeTypeCompound, true)
	if err != nil {
		return nil, fmt.Errorf("read nbt data: %w", err)
	}

	return &File{
		Root: rootNode,
	}, nil
}

func readRawByte(r io.Reader) (byte, error) {
	val := make([]byte, 1)
	if _, err := io.ReadFull(r, val); err != nil {
		return 0, err
	}
	return val[0], nil
}

func readRawUShort(r io.Reader) (uint16, error) {
	val := make([]byte, 2)
	if _, err := io.ReadFull(r, val); err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint16(val), nil
}

func readRawInt(r io.Reader) (int32, error) {
	val := make([]byte, 4)
	if _, err := io.ReadFull(r, val); err != nil {
		return 0, err
	}
	return int32(binary.BigEndian.Uint32(val)), nil
}

func readRawString(r io.Reader) (string, error) {
	strLen, err := readRawUShort(r)
	if err != nil {
		return "", err
	}
	val := make([]byte, strLen)
	if _, err := io.ReadFull(r, val); err != nil {
		return "", err
	}
	return string(val), nil
}

func readRawNodeType(r io.Reader) (NodeType, error) {
	val, err := readRawByte(r)
	if err != nil {
		return 0, err
	}
	return NodeType(val), nil
}

func readNode(r io.Reader) (Node, error) {
	nodeType, err := readRawNodeType(r)
	if err != nil {
		return nil, err
	}

	return readNodeOfType(r, nodeType, false)
}

func readNodeOfType(r io.Reader, nodeType NodeType, isRoot bool) (Node, error) {
	switch nodeType {
	case NodeTypeByte:
		return readByteNode(r)
	case NodeTypeShort:
		return readShortNode(r)
	case NodeTypeInt:
		return readIntNode(r)
	case NodeTypeLong:
		return readLongNode(r)
	case NodeTypeFloat:
		return readFloatNode(r)
	case NodeTypeDouble:
		return readDoubleNode(r)
	case NodeTypeString:
		return readStringNode(r)
	case NodeTypeList:
		return readListNode(r)
	case NodeTypeCompound:
		return readCompoundNode(r, isRoot)
	case NodeTypeIntArray:
		return readIntArrayNode(r)

	default:
		return nil, fmt.Errorf("unsupported node type %v", nodeType)
	}
}

type ByteNode struct {
	Value byte
}

func (n *ByteNode) Type() NodeType { return NodeTypeByte }

func readByteNode(r io.Reader) (*ByteNode, error) {
	val, err := readRawByte(r)
	if err != nil {
		return nil, err
	}
	return &ByteNode{
		Value: val,
	}, nil
}

type ShortNode struct {
	Value int16
}

func (n *ShortNode) Type() NodeType { return NodeTypeShort }

func readShortNode(r io.Reader) (*ShortNode, error) {
	val := make([]byte, 2)
	if _, err := io.ReadFull(r, val); err != nil {
		return nil, err
	}
	return &ShortNode{
		Value: int16(binary.BigEndian.Uint16(val)),
	}, nil
}

type IntNode struct {
	Value int32
}

func (n *IntNode) Type() NodeType { return NodeTypeInt }

func readIntNode(r io.Reader) (*IntNode, error) {
	val, err := readRawInt(r)
	if err != nil {
		return nil, err
	}
	return &IntNode{
		Value: val,
	}, nil
}

type LongNode struct {
	Value int64
}

func (n *LongNode) Type() NodeType { return NodeTypeLong }

func readLongNode(r io.Reader) (*LongNode, error) {
	val := make([]byte, 8)
	if _, err := io.ReadFull(r, val); err != nil {
		return nil, err
	}
	return &LongNode{
		Value: int64(binary.BigEndian.Uint64(val)),
	}, nil
}

type FloatNode struct {
	Value float32
}

func (n *FloatNode) Type() NodeType { return NodeTypeFloat }

func readFloatNode(r io.Reader) (*FloatNode, error) {
	val := make([]byte, 4)
	if _, err := io.ReadFull(r, val); err != nil {
		return nil, err
	}
	return &FloatNode{
		Value: math.Float32frombits(binary.BigEndian.Uint32(val)),
	}, nil
}

type DoubleNode struct {
	Value float64
}

func (n *DoubleNode) Type() NodeType { return NodeTypeDouble }

func readDoubleNode(r io.Reader) (*DoubleNode, error) {
	val := make([]byte, 8)
	if _, err := io.ReadFull(r, val); err != nil {
		return nil, err
	}
	return &DoubleNode{
		Value: math.Float64frombits(binary.BigEndian.Uint64(val)),
	}, nil
}

type StringNode struct {
	Value string
}

func (n *StringNode) Type() NodeType { return NodeTypeInt }

func readStringNode(r io.Reader) (*StringNode, error) {
	val, err := readRawString(r)
	if err != nil {
		return nil, err
	}
	return &StringNode{
		Value: val,
	}, nil
}

type ListNode struct {
	Values []Node
}

func (n *ListNode) Type() NodeType { return NodeTypeList }

func readListNode(r io.Reader) (*ListNode, error) {
	childNodeType, err := readRawNodeType(r)
	if err != nil {
		return nil, err
	}

	childCount, err := readRawInt(r)
	if err != nil {
		return nil, err
	}

	node := ListNode{
		Values: make([]Node, childCount),
	}
	for i := range int(childCount) {
		childNode, err := readNodeOfType(r, childNodeType, false)
		if err != nil {
			return nil, fmt.Errorf("read list index %d: %w", i, err)
		}

		node.Values = append(node.Values, childNode)
	}
	return &node, nil
}

type CompoundNode struct {
	Values map[string]Node
}

func (n *CompoundNode) Type() NodeType { return NodeTypeCompound }

func readCompoundNode(r io.Reader, isRoot bool) (*CompoundNode, error) {
	node := CompoundNode{
		Values: make(map[string]Node),
	}
	for {
		childNodeType, err := readRawNodeType(r)
		if err != nil {
			return nil, err
		}

		if childNodeType == NodeTypeEnd {
			break
		}

		childName, err := readRawString(r)
		if err != nil {
			return nil, err
		}
		fmt.Println(childName)

		childNode, err := readNodeOfType(r, childNodeType, false)
		if err != nil {
			return nil, fmt.Errorf("read compound child %q: %w", childName, err)
		}

		node.Values[childName] = childNode

		if isRoot {
			// the root-node only has a single value
			break
		}
	}
	return &node, nil
}

type IntArrayNode struct {
	Values []Node
}

func (n *IntArrayNode) Type() NodeType { return NodeTypeIntArray }

func readIntArrayNode(r io.Reader) (*IntArrayNode, error) {
	childCount, err := readRawInt(r)
	if err != nil {
		return nil, err
	}

	node := IntArrayNode{
		Values: make([]Node, childCount),
	}
	for i := range int(childCount) {
		childNode, err := readNodeOfType(r, NodeTypeInt, false)
		if err != nil {
			return nil, fmt.Errorf("read list index %d: %w", i, err)
		}

		node.Values = append(node.Values, childNode)
	}
	return &node, nil
}
