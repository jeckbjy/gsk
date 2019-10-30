package cli

//type RankCmdService struct {
//}
//
//type RankTopCmd struct {
//	Uid    string `cli:"0"`
//	RankId string `cli:"1"`
//}
//
//func (*RankCmdService) Top() {
//
//}
//
//// cli.Register(RankCmdService{})
//// cli.Register(RedeemCmdService{})
//
//type RedeemCmdService struct {
//}
//
//// redeem gen id 1 -t true -a aaa
//type RedeemGenCmd struct {
//	CouponID string `cli:"p,0"`
//	Count    int    `cli:"p,1"`
//	Test     bool   `cli:"f,,required,multiple"`
//	Account  string `cli:"f,account|a,"`
//}
//
//func (c *RedeemGenCmd) Run(ctx context.Context) {
//
//}
//
//func (*RedeemCmdService) Gen(ctx context.Context, args *RedeemGenCmd) {
//	//
//}
//
//// redeem download 1
//type RedeemDownloadCmd struct {
//	Id string `cli:"p,0"`
//}
//
//func (c *RedeemDownloadCmd) Run(ctx context.Context) {
//
//}
//
//func (*RedeemCmdService) Download(ctx context.Context, args *RedeemGenCmd) {
//
//}
