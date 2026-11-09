package messagestyle

import "testing"

func TestNormalizeAcceptsAndCanonicalizesStyle(t *testing.T){
	got,err:=(Style{FontFamily:" Tahoma ",FontSize:18,Color:"#aabbcc",Bold:true}).Normalize()
	if err!=nil{t.Fatalf("Normalize() error=%v",err)}
	if got.FontFamily!="Tahoma"||got.Color!="#AABBCC"||!got.Bold{t.Fatalf("unexpected style: %+v",got)}
}
func TestNormalizeRejectsInvalidValues(t *testing.T){
	tests:=[]Style{
		{FontFamily:"Comic Sans",FontSize:14,Color:"#FFFFFF"},
		{FontFamily:"Arial",FontSize:9,Color:"#FFFFFF"},
		{FontFamily:"Arial",FontSize:29,Color:"#FFFFFF"},
		{FontFamily:"Arial",FontSize:14,Color:"red"},
		{FontFamily:"Arial",FontSize:14,Color:"#FFF"},
	}
	for _,v:=range tests{if _,err:=v.Normalize();err==nil{t.Errorf("Normalize(%+v) succeeded",v)}}
}
func TestDefaultStyleIsValid(t *testing.T){if _,err:=DefaultStyle().Normalize();err!=nil{t.Fatalf("default style invalid: %v",err)}}
