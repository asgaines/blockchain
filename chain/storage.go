package chain

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	pb "github.com/asgaines/blockchain/protogo/blockchain"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
)

func InitChain(hasher Hasher, filesPrefix string) *Chain {
	b, err := ioutil.ReadFile(getStorageFnameProto(filesPrefix))
	if err != nil {
		return NewChain(hasher)
	}

	var bcpb pb.Chain

	if err := proto.Unmarshal(b, &bcpb); err != nil {
		log.Fatalf("could not unmarshal chain: %s", err)
	}

	bc := Chain{
		Pbc: &bcpb,
	}

	if !bc.IsSolid(hasher) {
		log.Fatalf("Initialization failed due to broken chain in storage file: %s", getStorageFnameProto(filesPrefix))
	}

	return &bc
}

func (c *Chain) Store(filesPrefix string) error {
	f, err := os.OpenFile(getStorageFnameProto(filesPrefix), os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Println(err)
		}
	}()

	b, err := proto.Marshal(c.ToProto())
	if err != nil {
		return fmt.Errorf("could not marshal chain: %w", err)
	}

	if _, err := f.Write(b); err != nil {
		return fmt.Errorf("could not write to file: %w", err)
	}

	fj, err := os.OpenFile(getStorageFnameJSON(filesPrefix), os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer func() {
		if err := fj.Close(); err != nil {
			log.Println(err)
		}
	}()

	if err := new(jsonpb.Marshaler).Marshal(fj, c.ToProto()); err != nil {
		return err
	}

	return nil
}

func getStorageFnameProto(filesPrefix string) string {
	return fmt.Sprintf("%s.proto", filesPrefix)
}

func getStorageFnameJSON(filesPrefix string) string {
	return fmt.Sprintf("%s.json", filesPrefix)
}
