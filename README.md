# plate-reader-poller

# 写在前面

本篇文章应边无际Edgenesis【Go后端开发实习生】的笔试要求而写。

# 1. 安装shifu

笔者的windows环境下已经有了wsl2和docker desktop，故直接命令行安装shifu即可。

```shell
curl -sfL https://raw.githubusercontent.com/Edgenesis/shifu/main/test/scripts/shifu-demo-install.sh | sudo sh -
```

安装完成后，在decker desktop中肉眼可见地多了很多镜像：

![image.png](https://github.com/ChenaLi0816/plate-reader-poller/blob/main/img/1.png)

及一个在运行的容器：

![image.png](https://github.com/ChenaLi0816/plate-reader-poller/blob/main/img/2.png)

nginx及其他设备正常运行：

![image.png](https://github.com/ChenaLi0816/plate-reader-poller/blob/main/img/3.png)

# 2. 运行酶标仪的数字孪生

创建酶标仪的数字孪生：

```shell
sudo kubectl apply -f run_dir/shifu/demo_device/edgedevice-plate-reader
```

获取酶标仪数据：
![image.png](https://github.com/ChenaLi0816/plate-reader-poller/blob/main/img/4.png)

```shell
curl "deviceshifu-plate-reader.deviceshifu.svc.cluster.local/get_measurement"
```

# 3. 编写go应用

1. 定期轮询获取酶标仪的/get_measurement接口
    - 思路：循环sleep一段时间或者设置一个定时器，定时向`http://deviceshifu-plate-reader.deviceshifu.svc.cluster.local/get_measurement`发送`GET`请求获取到数据
2. 将返回值平均后打印出来
    - 思路：将`response body`读出后逐个数字进行浮点数转化，相加后求其平均
3. 轮询时间可自定义
    - 思路：获取环境变量`poll_interval`的值作为轮询时间，单位为秒，如果不存在则默认10s

因为整体逻辑不是很复杂所以就全都写在`main.go`文件里了。

生成go可执行文件：

```shell
go build -o main
```

在临时的容器中调试运行的效果：

![image.png](https://github.com/ChenaLi0816/plate-reader-poller/blob/main/img/5.png)

# 4. 容器部署

## 镜像准备

写一个dockerfile：（这里的main是上一步生成的go可执行文件）

```dockerfike
FROM ubuntu:latest

WORKDIR /poller

COPY ./main /poller/

ENTRYPOINT ["./main"]
```

制作镜像：

```shell
docker build -t go-poller -f dockerfile .
```

之后在docker desktop会出现相应的镜像：

![image.png](https://github.com/ChenaLi0816/plate-reader-poller/blob/main/img/6.png)

打上标签

```shell
docker tag go-poller registry.cn-hangzhou.aliyuncs.com/shifu/go-poller
```

推送到镜像仓库：（这里用的是阿里云的个人版镜像仓库）

```shell
docker push registry.cn-hangzhou.aliyuncs.com/shifu/go-poller
```

![image.png](https://github.com/ChenaLi0816/plate-reader-poller/blob/main/img/7.png)

# 5. k8s编排

## 方法1：直接run

参考跑nginx的方法，键入命令：

```shell
sudo kubectl run --image=registry.cn-hangzhou.aliyuncs.com/shifu/go-poller mypoller
```

![image.png](https://github.com/ChenaLi0816/plate-reader-poller/blob/main/img/8.png)


`sudo kubectl get pods`查看是否运行：
![image.png](https://github.com/ChenaLi0816/plate-reader-poller/blob/main/img/9.png)
运行成功。

`sudo kubectl logs mypoller`查看日志：
![image.png](https://github.com/ChenaLi0816/plate-reader-poller/blob/main/img/10.png)
日志输出正常。

## (2)方法2：编写yaml文件

编写文件过程参考了[Kubernetes之yaml文件详解](https://www.cnblogs.com/lgeng/p/11053063.html)

配置文件`mypoller.yaml`如下：

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: mypoller
spec:
  containers:
  - image: registry.cn-hangzhou.aliyuncs.com/shifu/go-poller
    name: mypoller
    env:
    - name: poll_interval
      value: "15"
```

应用配置文件运行并查看日志：

```shell
sudo kubectl apply -f mypoller.yaml
sudo kubectl logs mypoller
```

![image.png](https://github.com/ChenaLi0816/plate-reader-poller/blob/main/img/11.png)
这里的轮询时间变为了15s一次，因为在`mypoller.yaml`文件内设置了容器的环境变量`poll_interval`。

**遇到的问题**

最开始时我在yaml文件里设置环境变量是这样写的：

```yaml
env: 
- name: poll_interval 
  value: 15
```

`apply`后报了错误：

```
Pod in version "v1" cannot be handled as a Pod: json: cannot unmarshal number into Go struct field EnvVar.spec.containers.env.value of type string
```

错误原因是说不能将数字反序列化为某个类型为`string`的字段，后来在数字`15`加上了双引号后解决。

# 总结

总体过程来说比较顺利，在`shifu`上完全没有遇到问题，这也得益于官方文档简单清晰。

go应用的思路也相对简单。

唯一比较卡住的点应该是在k8s的部署上，对于一些概念（`node`,`deployment`,`service`,`pod`)之前没有了解过所以需要查一些资料，好在k8s的命令和docker的命令也比较相似，之前有过docker部署经验所以学习k8s也不算艰难。