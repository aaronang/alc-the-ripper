package master

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/spf13/viper"
)

func init() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		panic("Viper did not read the config correctly")
	}
}

func TestCreateAndTerminateSlave(t *testing.T) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(viper.GetString("aws.region"))},
	)
	if err != nil {
		t.Error("Could not create AWS session.", err)
	}

	svc := ec2.New(sess)

	instance, err := CreateSlave(svc)
	if err != nil {
		t.Error("Could not create instance", err)
	}

	if _, err = TerminateSlave(svc, instance); err != nil {
		t.Error("Could not terminate instance", err)
	}
}
