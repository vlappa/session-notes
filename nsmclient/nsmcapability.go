package nsmclient

type nsmCapability string

func (c nsmCapability) String() string {
	return string(c)
}

type nsmServerCapability string

func (s nsmServerCapability) String() string {
	return string(s)
}
