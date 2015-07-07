package otr3

import (
	"io"
	"math/big"
	"testing"
)

func defaultRand() io.Reader {
	return fixedRand([]string{
		"ABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCD",
		"BBCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCD",
		"CBCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCD",
		"DBCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCD",
	})
}

func Test_generateSMPSecretGeneratesASecret(t *testing.T) {
	aliceFingerprint := hexToByte("0102030405060708090A0B0C0D0E0F1011121314")
	bobFingerprint := hexToByte("3132333435363738393A3B3C3D3E3F4041424344")
	ssid := hexToByte("FFF1D1E412345668")
	secret := []byte("this is something secret")
	result := generateSMPSecret(aliceFingerprint, bobFingerprint, ssid, secret)
	assertDeepEquals(t, result, hexToByte("D9B2E56321F9A9F8E364607C8C82DECD8E8E6209E2CB952C7E649620F5286FE3"))
}

func Test_generatesLongerAandRValuesForOtrV3(t *testing.T) {
	otr := context{otrV3{}, defaultRand()}
	smp := otr.generateSMPStartParameters()
	assertDeepEquals(t, smp.a2, new(big.Int).SetBytes(hexToByte("ABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCD")))
	assertDeepEquals(t, smp.a3, new(big.Int).SetBytes(hexToByte("BBCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCD")))
	assertDeepEquals(t, smp.r2, new(big.Int).SetBytes(hexToByte("CBCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCD")))
	assertDeepEquals(t, smp.r3, new(big.Int).SetBytes(hexToByte("DBCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCDABCD")))
}

func Test_generatesShorterAandRValuesForOtrV2(t *testing.T) {
	otr := context{otrV2{}, defaultRand()}
	smp := otr.generateSMPStartParameters()
	assertDeepEquals(t, smp.a2, new(big.Int).SetBytes(hexToByte("ABCDABCDABCDABCDABCDABCDABCDABCD")))
	assertDeepEquals(t, smp.a3, new(big.Int).SetBytes(hexToByte("BBCDABCDABCDABCDABCDABCDABCDABCD")))
	assertDeepEquals(t, smp.r2, new(big.Int).SetBytes(hexToByte("CBCDABCDABCDABCDABCDABCDABCDABCD")))
	assertDeepEquals(t, smp.r3, new(big.Int).SetBytes(hexToByte("DBCDABCDABCDABCDABCDABCDABCDABCD")))
}

func Test_computesG2aAndG3aCorrectlyForOtrV3(t *testing.T) {
	otr := context{otrV3{}, defaultRand()}
	smp := otr.generateSMPStartParameters()
	expected1, _ := new(big.Int).SetString("2403828132201280691996934790042280522039644351533698950188369222708084962201190259148331059300969970454760451858146292777271089101335555432831591932280897663128605540136047489902371842091500899696205003971442783016830569070154151538113509331728265831584544751982441886086856394142221987668411465357977595712909149546719860976278668856500931856381813557383455751167139453911498547553424315774735908660702360108350494792398612100509843726869895575665079703048760466", 10)
	expected2, _ := new(big.Int).SetString("545300291901380653980556916625957052030430375409433663251760516986016273974200122997230389944238471404021073102701550317663669624982742390354806536963349739870183347869815633451836184053254713103530928305206048828718330259320999178725064437137499780326636715585953004481720026913428336861587141398794337053489567564530999611735085631734084309755871299251468595807420396099483025772958221192002020467585351536202731371781577659533010401568100631459405229307956337", 10)
	assertDeepEquals(t, smp.msg1.g2a, expected1)
	assertDeepEquals(t, smp.msg1.g3a, expected2)
}

func Test_computesG2aAndG3aCorrectlyForOtrV2(t *testing.T) {
	otr := context{otrV2{}, defaultRand()}
	smp := otr.generateSMPStartParameters()
	expected1, _ := new(big.Int).SetString("8a88c345c63aa25dab9815f8c51f6b7b621a12d31c8220a0579381c1e2e85a2275e2407c79c8e6e1f72ae765804e6b4562ac1b2d634313c70d59752ac119c6da5cb95dde3eedd9c48595b37256f5b64c56fb938eb1131447c9af9054b42841c57d1f41fe5aa510e2bd2965434f46dd0473c60d6114da088c7047760b00bc10287a03afc4c4f30e1c7dd7c9dbd51bdbd049eb2b8921cbdc72b4f69309f61e559c2d6dec9c9ce6f38ccb4dfd07f4cf2cf6e76279b88b297848c473e13f091a0f77", 16)
	expected2, _ := new(big.Int).SetString("d275468351fd48246e406ee74a8dc3db6ee335067bfa63300ce6a23867a1b2beddbdae9a8a36555fd4837f3ef8bad4f7fd5d7b4f346d7c7b7cb64bd7707eeb515902c66aa0c9323931364471ab93dd315f65c6624c956d74680863a9388cd5d89f1b5033b1cf232b8b6dcffaaea195de4e17cc1ba4c99497be18c011b2ad7742b43fa9ee3f95f7b6da02c8e894d054eb178a7822273655dc286ad15874687fe6671908d83662e7a529744ce4ea8dad49290d19dbe6caba202a825a20a27ee98a", 16)
	assertDeepEquals(t, smp.msg1.g2a, expected1)
	assertDeepEquals(t, smp.msg1.g3a, expected2)
}

func Test_computesC2AndD2CorrectlyForOtrV2(t *testing.T) {
	otr := context{otrV2{}, defaultRand()}
	smp := otr.generateSMPStartParameters()
	expected1, _ := new(big.Int).SetString("d3b6ef5528fa97e983395bec165fa4ced7657bdabf3742d60880965c369c880c", 16)
	expected2, _ := new(big.Int).SetString("7fffffffffffffffe487ed5110b4611a62633145c06e0e68948127044533e63a0105df531d89cd9128a5043cc71a026ef7ca8cd9e69d218d98158536f92f8a1ba7f09ab6b6a8e122f242dabb312f3f637a262174d31bf6b585ffae5b7a035bf6f71c35fdad44cfd2d74f9208be258ff324943328f6722d9ee1003e5c50b1df82cc6d241b0e2ae9cd348b1fd47e9267af339d65211b4fcfa466656c89b4217f90102e4aa3ac176a41f6240f32689712b0391c1c659757f4bfb83e6ba66bf8b630", 16)
	assertDeepEquals(t, smp.msg1.c2, expected1)
	assertDeepEquals(t, smp.msg1.d2, expected2)
}

func Test_computesC3AndD3CorrectlyForOtrV2(t *testing.T) {
	otr := context{otrV2{}, defaultRand()}
	smp := otr.generateSMPStartParameters()
	expected1, _ := new(big.Int).SetString("57d8cfda442854ecb01b28e631aa9165d51d1192f7f464bf17ea7f6665c05030", 16)
	expected2, _ := new(big.Int).SetString("7fffffffffffffffe487ed5110b4611a62633145c06e0e68948127044533e63a0105df531d89cd9128a5043cc71a026ef7ca8cd9e69d218d98158536f92f8a1ba7f09ab6b6a8e122f242dabb312f3f637a262174d31bf6b585ffae5b7a035bf6f71c35fdad44cfd2d74f9208be258ff324943328f6722d9ee1003e5c50b1df82cc6d241b0e2ae9cd348b1fd47e9267af8140bb2aa65628bcff455920bba95a1392f2fcb5c115f43a7a828b5bf0393c5c775a17a88506a7893ff509d674cd655c", 16)
	assertDeepEquals(t, smp.msg1.c3, expected1)
	assertDeepEquals(t, smp.msg1.d3, expected2)
}

// func Test_thatVerifySMPStartParametersCheckG2AForOtrV3(t *testing.T) {
// 	otr := context{otrV3{}, defaultRand()}
// 	otr.verifySMPStartParameters()
// }
