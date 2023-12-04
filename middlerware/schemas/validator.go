package schemas

var taskValidatorErrorMessage = map[string]string{
	"Domainsrequired": "缺少任务目标",
}

// RegisterValidatorRule 注册参数验证错误消息, Key = e.StructNamespace(), value.key = e.Field()+e.Tag()
var RegisterValidatorRule = map[string]map[string]string{
	"DomainParams": taskValidatorErrorMessage,
}
