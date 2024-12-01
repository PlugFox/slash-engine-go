// Code generated by the FlatBuffers compiler. DO NOT EDIT.

package Game

import (
	flatbuffers "github.com/google/flatbuffers/go"
)

type ImpulseT struct {
	Direction *VectorT
	Damping float64
	Next *ImpulseT
}

func (t *ImpulseT) Pack(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	if t == nil { return 0 }
	NextOffset := t.Next.Pack(builder)
	ImpulseStart(builder)
	DirectionOffset := t.Direction.Pack(builder)
	ImpulseAddDirection(builder, DirectionOffset)
	ImpulseAddDamping(builder, t.Damping)
	ImpulseAddNext(builder, NextOffset)
	return ImpulseEnd(builder)
}

func (rcv *Impulse) UnPackTo(t *ImpulseT) {
	t.Direction = rcv.Direction(nil).UnPack()
	t.Damping = rcv.Damping()
	t.Next = rcv.Next(nil).UnPack()
}

func (rcv *Impulse) UnPack() *ImpulseT {
	if rcv == nil { return nil }
	t := &ImpulseT{}
	rcv.UnPackTo(t)
	return t
}

type Impulse struct {
	_tab flatbuffers.Table
}

func GetRootAsImpulse(buf []byte, offset flatbuffers.UOffsetT) *Impulse {
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &Impulse{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *Impulse) Init(buf []byte, i flatbuffers.UOffsetT) {
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *Impulse) Table() flatbuffers.Table {
	return rcv._tab
}

func (rcv *Impulse) Direction(obj *Vector) *Vector {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		x := o + rcv._tab.Pos
		if obj == nil {
			obj = new(Vector)
		}
		obj.Init(rcv._tab.Bytes, x)
		return obj
	}
	return nil
}

func (rcv *Impulse) Damping() float64 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		return rcv._tab.GetFloat64(o + rcv._tab.Pos)
	}
	return 0.0
}

func (rcv *Impulse) MutateDamping(n float64) bool {
	return rcv._tab.MutateFloat64Slot(6, n)
}

func (rcv *Impulse) Next(obj *Impulse) *Impulse {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(8))
	if o != 0 {
		x := rcv._tab.Indirect(o + rcv._tab.Pos)
		if obj == nil {
			obj = new(Impulse)
		}
		obj.Init(rcv._tab.Bytes, x)
		return obj
	}
	return nil
}

func ImpulseStart(builder *flatbuffers.Builder) {
	builder.StartObject(3)
}
func ImpulseAddDirection(builder *flatbuffers.Builder, Direction flatbuffers.UOffsetT) {
	builder.PrependStructSlot(0, flatbuffers.UOffsetT(Direction), 0)
}
func ImpulseAddDamping(builder *flatbuffers.Builder, Damping float64) {
	builder.PrependFloat64Slot(1, Damping, 0.0)
}
func ImpulseAddNext(builder *flatbuffers.Builder, Next flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(2, flatbuffers.UOffsetT(Next), 0)
}
func ImpulseEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	return builder.EndObject()
}
