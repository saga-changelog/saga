// Package courier defines the contract between saga and its courier plugins,
// and provides the runtime harness that courier implementations use to wire
// themselves up as CLI binaries.
//
// A courier is an external binary, named saga-courier-<name>, that delivers
// tales to a specific platform. Writing one consists of implementing the
// Courier interface and calling Run from main():
//
//	type slackCourier struct{}
//
//	func (slackCourier) Info() courier.Info { ... }
//	func (slackCourier) ValidateConfig(ctx context.Context, p *courier.Payload) error { ... }
//	func (slackCourier) Tell(ctx context.Context, p *courier.Payload) error { ... }
//
//	func main() { courier.Run(slackCourier{}) }
package courier
