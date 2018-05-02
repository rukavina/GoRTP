package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"

	"github.com/rukavina/GoRTP/pkg/rtp"
)

var localPort = 4000
var local, _ = net.ResolveIPAddr("ip", "127.0.0.1")

var remotePort = 5000
var remote, _ = net.ResolveIPAddr("ip", "127.0.0.1")

var rsLocal *rtp.Session

func main() {
	tpLocal, _ := rtp.NewTransportUDP(local, localPort, local.Zone)
	rsLocal = rtp.NewSession(tpLocal, tpLocal)

	// Add address of a remote peer (participant)
	_, err := rsLocal.AddRemote(&rtp.Address{remote.IP, remotePort, remotePort + 1, remote.Zone})
	if err != nil {
		log.Fatalf("Error adding remote: %s", err)
	}

	// Create a media stream.
	// The SSRC identifies the stream. Each stream has its own sequence number and other
	// context. A RTP session can have several RTP stream for example to send several
	// streams of the same media.
	//
	strLocalIdx, _ := rsLocal.NewSsrcStreamOut(&rtp.Address{local.IP, localPort, localPort + 1, local.Zone}, 1020304, 4711)
	rsLocal.SsrcStreamOutForIndex(strLocalIdx).SetPayloadType(0)

	rsLocal.StartSession()

	go sendLocalToRemote()

	time.Sleep(8e9)

	time.Sleep(30e6) // allow the sender to drain

	rsLocal.CloseSession()

	time.Sleep(10e6)

	fmt.Println("Streaming file done")
}

// Create a RTP packet suitable for standard stream (index 0) with a payload length of 160 bytes
// The method initializes the RTP packet with SSRC, sequence number, and RTP version number.
// If the payload type was set with the RTP stream then the payload type is also set in
// the RTP packet
func sendLocalToRemote() {
	stamp := uint32(0)

	file, err := os.Open("welcome-ulaw.wav")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	header := make([]byte, 58)
	_, err = file.Read(header)

	buffer := make([]byte, 160)
	for {
		bytesread, err := file.Read(buffer)
		if err != nil {
			if err != io.EOF {
				fmt.Println(err)
			}
			break
		}
		log.Printf("bytes read from file: %d", bytesread)

		rp := rsLocal.NewDataPacket(stamp)
		rp.SetPayload(buffer[:bytesread])
		bytessent, err := rsLocal.WriteData(rp)
		if err != nil {
			log.Printf("Error sending remote data: %s", err)
		} else {
			log.Printf("bytes sent to remote peer: %d", bytessent)
		}
		rp.FreePacket()

		stamp += 160
		time.Sleep(20e6)
	}
}
