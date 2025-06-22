## 更新日志
**v0.0.133**
增加对OpenKruise的支持：
advancedcronjobs.apps.kruise.io                
broadcastjobs.apps.kruise.io                   
clonesets.apps.kruise.io                       
containerrecreaterequests.apps.kruise.io       
daemonsets.apps.kruise.io                      
imagelistpulljobs.apps.kruise.io               
imagepulljobs.apps.kruise.io                   
nodeimages.apps.kruise.io                      
nodepodprobes.apps.kruise.io                   
persistentpodstates.apps.kruise.io             
podprobemarkers.apps.kruise.io                 
podunavailablebudgets.policy.kruise.io         
resourcedistributions.apps.kruise.io           
sidecarsets.apps.kruise.io                     
statefulsets.apps.kruise.io                    
uniteddeployments.apps.kruise.io               
workloadspreads.apps.kruise.io                 

**v0.0.132**
支持用户禁用，禁用后不可登录、不可API访问、不可MCP访问

**v0.0.130**
增加自定义规则巡检功能。
支持使用lua脚本编写自定义规则。
支持设置webhook接收通知。
支持AI总结巡检结果。 

**v0.0.128**
新增命名空间黑白名单管控
白名单命名空间：置空表示不限制，可访问该集群下所有的命名空间。如果填写了，那么用户就只能访问指定的命名空间了。
黑名单命名空间：置空表示不限制，如果填写了，那么用户将不能访问该命名空间。黑名单可否定白名单。黑名单权限最高。


**v0.0.124**
新增模型配置管理，可以增加多个大模型配置。

**v0.0.123**
本版本优化了内存占用

**v0.0.120**
Fix 修复用户组权限类型错误问题
feat(AI): 添加关闭AI思考过程输出的功能

**v0.0.119**
新增Mysql、Postgresql数据库支持

**v0.0.118**
Fix shell 等待超时问题

**v0.0.117**
新增权限缓存，降低sql压力

**v0.0.116**
修复大模型对话记录跟其他单轮解释互相影响的问题

**v0.0.115**
Fix probe 探测失败导致无法访问的问题

**v0.0.114**
新增Github Copilot MCP 功能支持

**v0.0.113**
MCP 访问路径兼容动态路径/mcp/k8m/xxxx/sse格式，适用于不支持设置Header的MCP客户端

**v0.0.112**
新增Github Copilot MCP 功能支持

**v0.0.111**
新增大模型保存yaml为模板功能。聊天过程中，将大模型给出的yaml示例，保存为模板。
在资源新建页面，可以选择这个模板直接使用。

**v0.0.108**
AI聊天窗口增加清空历史记录按钮。

**v0.0.107**
新增AI多轮对话功能。可由大模型在对话中进行规划，并按步骤执行MCP工具，完成目标。

**v0.0.106**
新增端口转发功能

1. ![输入图片说明](https://foruda.gitee.com/images/1746886285039073315/cd86c8c8_77493.png "在这里输入图片标题")
2. ![输入图片说明](https://foruda.gitee.com/images/1746886307306899637/c6bae855_77493.png "iShot_2025-05-10_22.09.49.png")
3. 新增集群命名空间列表接口并优化命名空间选择功能

**v0.0.96更新**

1. 修复MCP调用记录问题
2. 调整新增deploy标签功能
3. 优化开放MCP调用页面显示完整访问路径
4. 新增拉取镜像超时时间配置
5. 新增指标等资源缓存超时时间

**v0.0.93更新**
新增Pod列表、节点列表、Ns列表页面的实时用量展示。

1. Pod列表展示效果如下：
   ![Image](https://github.com/user-attachments/assets/41e0283f-aa6e-432d-a62a-f1d142359929)
2. 节点展示效果如下：
   ![Image](https://github.com/user-attachments/assets/f67b1439-47e7-453e-ba85-51f4715b6bc4)
3. 命名空间展示效果如下：
   ![Image](https://github.com/user-attachments/assets/82e42baf-d688-4376-bbe1-97dde361b9e5)

**v0.0.92更新**

* 新增OIDC单点登录支持

![输入图片说明](https://foruda.gitee.com/images/1745080546954797409/3845a390_77493.png "屏幕截图")
![输入图片说明](https://foruda.gitee.com/images/1745080595835207078/ed115dc8_77493.png "屏幕截图")

**v0.0.88更新**

1. 新增MCP执行记录
   ![输入图片说明](https://foruda.gitee.com/images/1744644690534249581/dffdd2b2_77493.png "屏幕截图")
   ![输入图片说明](https://foruda.gitee.com/images/1744644698183429067/6fb8a635_77493.png "屏幕截图")
1. 新增集群名称前置检测
   在MCP工具执行前，检测集群名称是否正确。
   {
   "tool_name": "restart_deployment@k8m",
   "parameters": {
   "cluster": "config/kind-kind-cluster123",
   "name": "k8m",
   "namespace": "k8m"
   },
   "result": "",
   "error": "工具执行失败: cluster config/kind-kind-cluster123 not found 集群不存在，请检查集群名称"
   }
1. 修复Pod文件上传下载功能

**v0.0.87更新**

1. 集群授权支持对用户组进行授权
   集群授权：
   ![输入图片说明](https://foruda.gitee.com/images/1744554238488470925/351bbc00_77493.png "屏幕截图")
   用户管理视角，看用户有哪些集群权限：
   ![](https://foruda.gitee.com/images/1744554316927031816/24a3c6ce_77493.png "屏幕截图")
   集群管理视角，看某集群下已授权用户：
   ![输入图片说明](https://foruda.gitee.com/images/1744554384827407363/e3d0136b_77493.png "屏幕截图")
   用户视角，看自己有哪些已获得授权的集群列表：
   ![输入图片说明](https://foruda.gitee.com/images/1744554435367667674/1af1bd5e_77493.png "屏幕截图")

**v0.0.86更新**

1. 资源状态翻转
   新增状态指标翻转，将压力、问题等表述的状态，翻转显示为正常
   ![输入图片说明](https://foruda.gitee.com/images/1744466360319112414/5554605f_77493.png "屏幕截图")
   ![输入图片说明](https://foruda.gitee.com/images/1744466344472247290/484335b8_77493.png "屏幕截图")
2. 新增MCP工具的独立开关
   约束每一个工具，控制大模型可使用tools的范围，屏蔽高危操作，减低大模型交互负担。
   ![输入图片说明](https://foruda.gitee.com/images/1744466440407504939/108fd6d9_77493.png "屏幕截图")
3. 新增临时管理员账户配置开关
   开启后，可通过启动参数、环境变量设置平台管理员用户名密码。增加正常管理员后，可关闭临时管理员。
   该功能默认不生效，也就是不设置开启，只能使用数据库用户名密码登录。确保安全。
   建议生产环境非必要不要启用。
4. 新增集群自动连接开关
   开启后，会自动连接已注册的集群。

**v0.0.75更新**

1. 分离用户操作界面、平台管理界面。平台管理界面新增一个平台管理菜单。
   1.1 用户多集群切换，保留切换、连接功能:
   ![输入图片说明](https://foruda.gitee.com/images/1743904007097350906/c1dd8712_77493.png "屏幕截图")
   1.2 管理员操作多集群，新增断开功能：
   ![输入图片说明](https://foruda.gitee.com/images/1743904225916002465/6ee9a422_77493.png "屏幕截图")
   1.3 集群管理新增已授权页面，展示集群下所有的授权用户
   ![输入图片说明](https://foruda.gitee.com/images/1743904287185877723/dbc711cb_77493.png "屏幕截图")
   1.4 用户管理新增授权页面，查看某用户所有的授权集群
   ![输入图片说明](https://foruda.gitee.com/images/1743904361656769506/de632dca_77493.png "屏幕截图")
2. 新增权限可设置ns，集群授权后，可补充ns，默认为不限制，填写后，将限制用户活动范围。
   ![输入图片说明](https://foruda.gitee.com/images/1743904016110156134/7aa4c81c_77493.png "屏幕截图")
   ![输入图片说明](https://foruda.gitee.com/images/1743904052295697800/8f38845c_77493.png "屏幕截图")
3. 新增参数配置页面。
   启动后会先加载环境变量、env文件、页面配置，依次覆盖。最终页面配置为准。
   ![输入图片说明](https://foruda.gitee.com/images/1743904079152543105/cf923008_77493.png "屏幕截图")
4. 新增资源、副本数调整页面
   ![输入图片说明](https://foruda.gitee.com/images/1743904476260674721/310b0f04_77493.png "屏幕截图")
   ![输入图片说明](https://foruda.gitee.com/images/1743904500139794748/df5c4bed_77493.png "屏幕截图")

**V0.0.73 更新**

1. 新增Deploy探针管理页面
   ![输入图片说明](https://foruda.gitee.com/images/1743686996148531876/a3dc1131_77493.png "屏幕截图")
1. 新增MCP多集群不传值时提示，只有一个集群时可以省去集群名称
   ![输入图片说明](https://foruda.gitee.com/images/1743687014148809259/e5526f1f_77493.png "屏幕截图")
1. 修复未授权用户看到一个默认集群的问题

**V0.0.72 更新**

1. MCP 大模型调用权限上线，一句话概述：谁使用大模型，就用谁的权限执行MCP
   ![输入图片说明](https://foruda.gitee.com/images/1743650492231083539/72855c43_77493.png "屏幕截图")

**V0.0.70 更新**

1. 权限管理调整：按集群进行权限隔离
   ![输入图片说明](https://foruda.gitee.com/images/1743436163730546653/203d33f7_77493.png "屏幕截图")

**v0.0.67 更新**

1. 新增：MCP查询事件工具
   ![输入图片说明](https://foruda.gitee.com/images/1742916865442166281/43b26650_77493.png "屏幕截图")
2. 新增：MCP查询注册集群工具
   ![输入图片说明](https://foruda.gitee.com/images/1742917222171687147/216d03f1_77493.png "屏幕截图")
3. 新增：MCP查询事件工具
   ![输入图片说明](https://foruda.gitee.com/images/1742917268538391635/9e25fbb3_77493.png "屏幕截图")
4. 增强：列表查询资源支持label ，如app=k8m
   ![输入图片说明](https://foruda.gitee.com/images/1742916917319897798/a2171fd2_77493.png "屏幕截图")
5. 增强：MCP服务器增加快捷开启关闭按钮
   ![输入图片说明](https://foruda.gitee.com/images/1742916947056442916/6c33d7c2_77493.png "屏幕截图")

**V0.0.66更新**

1. 新增MCP支持。
2. 内置支持k8s多集群操作：
    1. list_k8s_resource
    2. get_k8s_resource
    3. delete_k8s_resource
    4. describe_k8s_resource
    5. get_pod_logs

**v0.0.64 更新**

1. 增加MCP支持
   ![输入图片说明](https://foruda.gitee.com/images/1742621225108846936/0a614dcb_77493.png "屏幕截图")
   ![输入图片说明](https://foruda.gitee.com/images/1742621196785322998/4174b937_77493.png "屏幕截图")
   ![输入图片说明](https://foruda.gitee.com/images/1742621204002335466/8a02cd2c_77493.png "屏幕截图")

**v0.0.62 更新**

1. 划词解释增加全屏按钮
   解决部分情况下解释内容非常多，查看不方便，以及滚动条不能完整滚动的问题。
   ![输入图片说明](https://foruda.gitee.com/images/1742085361623662812/c569323a_77493.png "屏幕截图")
   ![输入图片说明](https://foruda.gitee.com/images/1742085379102268742/769429f2_77493.png "屏幕截图")

**v0.0.61 更新**

1. 新增2FA两步验证
   启用后，登录时需填写验证码，增强安全性
   ![输入图片说明](https://foruda.gitee.com/images/1742012358386285979/eada8b94_77493.png "屏幕截图")
2. InCluster运行模式增加开关
   默认开启，可设置环境变量显式关闭。按需开启。
3. 优化资源用量显示逻辑
   未设置资源用量，在k8s中属于最低保障等级。界面显示进度条调整为红色100%，提醒管理员关注。
   ![资源用量](https://foruda.gitee.com/images/1742012525046823733/35acfc96_77493.png "屏幕截图")

**v0.0.60更新**

1. 增加helm 常用仓库
   ![输入图片说明](https://foruda.gitee.com/images/1741792802066909841/f20b8736_77493.png "屏幕截图")
   ![输入图片说明](https://foruda.gitee.com/images/1741792815487933294/a4b9c193_77493.png "屏幕截图")
2. Namespace增加LimitRange、ResourceQuota快捷菜单
   ![输入图片说明](https://foruda.gitee.com/images/1741792871141287157/f0a51266_77493.png "屏幕截图")
   ![输入图片说明](https://foruda.gitee.com/images/1741792891386812848/ad928eb1_77493.png "屏幕截图")
3. 增加InCluster模式开关
   默认开启InCluster模式，如需关闭，可以注入环境变量，或修改配置文件，或修改命令行参数

**v0.0.53更新**

1. 日志查看支持颜色，如果输出console的时候带有颜色，那么在pod 日志查看时就可以显示。
   ![输入图片说明](https://foruda.gitee.com/images/1741180128542917712/d4034cfb_77493.png "屏幕截图")
2. Helm功能上线
   2.1 新增helm仓库
   ![输入图片说明](https://foruda.gitee.com/images/1741180306318265893/f7c561cf_77493.png "屏幕截图")
   2.2 安装helm chart 应用
   应用列表
   ![输入图片说明](https://foruda.gitee.com/images/1741180337250117323/373632c3_77493.png "屏幕截图")
   查看应用
   ![输入图片说明](https://foruda.gitee.com/images/1741180373708023891/01b2eef5_77493.png "屏幕截图")
   ![输入图片说明](https://foruda.gitee.com/images/1741180423218217871/b1b2b06f_77493.png "屏幕截图")
   支持对参数内容选中划词AI解释
   ![输入图片说明](https://foruda.gitee.com/images/1741180604109610379/b26ae294_77493.png "屏幕截图")
   2.3 查看已部署release
   ![输入图片说明](https://foruda.gitee.com/images/1741180730249955448/bd51776e_77493.png "屏幕截图")
   ![输入图片说明](https://foruda.gitee.com/images/1741180757526613636/3cff8334_77493.png "屏幕截图")
   2.4 查看安装参数
   ![输入图片说明](https://foruda.gitee.com/images/1741180785289693466/dd1e08ab_77493.png "屏幕截图")
   2.5 更新、升级、降级部署版本
   ![输入图片说明](https://foruda.gitee.com/images/1741180817303995346/b2bb7472_77493.png "屏幕截图")
   2.6 查看已部署release变更历史
   ![输入图片说明](https://foruda.gitee.com/images/1741180840762812700/ccd3aa07_77493.png "屏幕截图")

**v0.0.50更新**

1. 新增HPA
   ![输入图片说明](https://foruda.gitee.com/images/1740664600490309267/48ff3895_77493.png "屏幕截图")
2. 关联资源增加HPA
   ![输入图片说明](https://foruda.gitee.com/images/1740664626159889748/96a40af4_77493.png "屏幕截图")

**v0.0.49更新**

1. 新增标签搜索：支持精确搜索、模糊搜索。
   精确搜索。可以搜索k，k=v两种方式精确搜索。默认列出所有标签。支持自定义新增搜索标签。
   ![输入图片说明](https://foruda.gitee.com/images/1740664804869894211/257140ad_77493.png "屏幕截图")
   模糊搜索。可以搜索k，v中的任意满足。类似like %xx%的搜索方式。
   ![输入图片说明](https://foruda.gitee.com/images/1740664820221541385/cf840a61_77493.png "屏幕截图")
2. 多集群纳管支持自定义名称。
   ![输入图片说明](https://foruda.gitee.com/images/1740664838997975455/95aeec37_77493.png "屏幕截图")
   ![输入图片说明](https://foruda.gitee.com/images/1740664855863544600/3496c16f_77493.png "屏幕截图")
3. 优化Pod状态显示
   在列表页展示pod状态，不同颜色区分正常运行与未就绪运行。
   ![输入图片说明](https://foruda.gitee.com/images/1740664869098640512/0d4002eb_77493.png "屏幕截图")
   ![输入图片说明](https://foruda.gitee.com/images/1740664883842793338/17f94df3_77493.png "屏幕截图")

**v0.0.44更新**

1. 新增kubectl shell 功能
   可以web 页面执行 kubectl 命令了
   ![输入图片说明](https://foruda.gitee.com/images/1740031049224924895/c8d5357b_77493.png "屏幕截图")
   ![输入图片说明](https://foruda.gitee.com/images/1740031092919251676/61e6246c_77493.png "屏幕截图")

2. 新增节点终端NodeShell
   在节点上执行命令
   ![输入图片说明](https://foruda.gitee.com/images/1740031147702527911/4cef40dc_77493.png "屏幕截图")
   ![输入图片说明](https://foruda.gitee.com/images/1740031249763550505/69fddee6_77493.png "屏幕截图")
3. 新增创建功能页面
   执行过的yaml会保存下来，下次打开页面可以直接点击，收藏的yaml可以导入导出。导出的文件为yaml，可以复用
   ![输入图片说明](https://foruda.gitee.com/images/1740031367996726581/e1a357b7_77493.png "屏幕截图")
   ![](https://foruda.gitee.com/images/1740031382494497806/d16b1a79_77493.png "屏幕截图")
   ![输入图片说明](https://foruda.gitee.com/images/1740031533749791121/4e64e286_77493.png "屏幕截图")
4. deploy、ds、sts等类型新增关联资源
   4.1 容器组
   直接显示其下受控的pod容器组，并提供快捷操作
   ![输入图片说明](https://foruda.gitee.com/images/1740031610441749272/cd485e87_77493.png "屏幕截图")
   4.2 关联事件
   显示deploy、rs、pod等所有相关的事件，一个页面看全相关事件
   ![deploy](https://foruda.gitee.com/images/1740031712446573977/320c920b_77493.png "屏幕截图")
   4.3 日志
   显示Pod列表，可选择某个pod、Container展示日志
   ![](https://foruda.gitee.com/images/1740031809856930240/fbbef393_77493.png "屏幕截图")
   4.4 历史版本
   支持历史版本查看，并可diff
   ![输入图片说明](https://foruda.gitee.com/images/1740031862075460381/ebf50a7e_77493.png "屏幕截图")
   ![输入图片说明](https://foruda.gitee.com/images/1740031912370086873/dfa95a2f_77493.png "屏幕截图")

5. 全新AI对话窗口
   ![输入图片说明](https://foruda.gitee.com/images/1740062818194113045/6ae3af0b_77493.png "屏幕截图")
   ![输入图片说明](https://foruda.gitee.com/images/1740062840392675452/a429aab8_77493.png "屏幕截图")

6. 全新AI搜索方式，哪里不懂选哪里
   页面所有地方都可以`划词翻译`,哪里有疑问就选中哪里。
   ![输入图片说明](https://foruda.gitee.com/images/1740062958174067230/7c377b16_77493.png "屏幕截图")

**v0.0.21更新**

1. 新增问AI功能：
   有什么问题，都可以直接询问AI，让AI解答你的疑惑
   ![输入图片说明](https://foruda.gitee.com/images/1736655942078335649/be66c2b5_77493.png "屏幕截图")
   ![输入图片说明](https://foruda.gitee.com/images/1736655968296155521/d47d247e_77493.png "屏幕截图")
2. 文档界面优化：
   优化AI翻译效果，降低等待时间
   ![AI文档](https://foruda.gitee.com/images/1736656055530922469/df155262_77493.png "屏幕截图")
3. 文档字段级AI示例：
   针对具体的字段，给出解释，给出使用Demo样例。
   ![输入图片说明](https://foruda.gitee.com/images/1736656231132357556/b41109e6_77493.png "屏幕截图")
4. 增加容忍度详情：
   ![输入图片说明](https://foruda.gitee.com/images/1736656289098443083/ce1f5615_77493.png "屏幕截图")
5. 增加Pod关联资源
   一个页面，展示相关的svc、endpoint、pvc、env、cm、secret，甚至集成了pod内的env列表，方便查看
   ![输入图片说明](https://foruda.gitee.com/images/1736656365325777082/410d24c5_77493.png "屏幕截图")
   ![输入图片说明](https://foruda.gitee.com/images/1736656376791203135/64cc4737_77493.png "屏幕截图")
   ![输入图片说明](https://foruda.gitee.com/images/1736656390371435096/5d93c74a_77493.png "屏幕截图")
   ![输入图片说明](https://foruda.gitee.com/images/1736656418411787086/2c8510af_77493.png "屏幕截图")
   ![输入图片说明](https://foruda.gitee.com/images/1736656445050779433/843f56aa_77493.png "屏幕截图")
   ![输入图片说明](https://foruda.gitee.com/images/1736656457940557219/c1372abd_77493.png "屏幕截图")
   ![输入图片说明](https://foruda.gitee.com/images/1736656468351816442/aba6f649_77493.png "屏幕截图")
6. yaml创建增加导入功能：
   增加导入功能，可以直接执行，也可导入到编辑器。导入编辑器后可以二次编辑后，再执行。
   ![输入图片说明](https://foruda.gitee.com/images/1736656627742328659/6c4e745e_77493.png "屏幕截图")
   ![输入图片说明](https://foruda.gitee.com/images/1736656647758880134/ca92dcc2_77493.png "屏幕截图")

**v0.0.19更新**

1. 多集群管理功能
   按需选择多集群，可随时切换集群
   ![输入图片说明](https://foruda.gitee.com/images/1736037285365941737/543965e6_77493.png "屏幕截图")
2. 节点资源用量功能
   直观显示已分配资源情况，包括cpu、内存、pod数量、IP数量。
   ![输入图片说明](https://foruda.gitee.com/images/1736037259029155963/72ea1ab4_77493.png "屏幕截图")
3. Pod 资源用量
   ![输入图片说明](https://foruda.gitee.com/images/1736037328973160586/9d322e6d_77493.png "屏幕截图")
4. Pod CPU内存设置
   按范围方式显示CPU设置，内存设置，简洁明了
   ![内存](https://foruda.gitee.com/images/1736037370125604986/7938a1f6_77493.png "屏幕截图")
5. AI页面功能升级为打字机效果
   响应速度大大提升，实时输出AI返回内容，体验升级
   ![输入图片说明](https://foruda.gitee.com/images/1736037522633946187/71955026_77493.png "屏幕截图")

**v0.0.15更新**

1. 所有页面增加资源使用指南。启用AI信息聚合。包括资源说明、使用场景（举例说明）、最佳实践、典型示例（配合前面的场景举例，编写带有中文注释的yaml示例）、关键字段及其含义、常见问题、官方文档链接、引用文档链接等信息，帮助用户理解k8s
   ![输入图片说明](https://foruda.gitee.com/images/1735400167081694530/e45b55ef_77493.png "屏幕截图")
2. 所有资源页面增加搜索功能。部分页面增高频过滤字段搜索。
   ![输入图片说明](https://foruda.gitee.com/images/1735399974060039020/11bce030_77493.png "屏幕截图")
3. 改进LimitRange信息展示模式
   ![LimitRange](https://foruda.gitee.com/images/1735399148267940416/b4faafbd_77493.png "屏幕截图")
4. 改进状态显示样式
   ![Deployment](https://foruda.gitee.com/images/1735399222088964660/131eda03_77493.png "屏幕截图")
5. 统一操作菜单
   ![操作菜单](https://foruda.gitee.com/images/1735399278081665887/b01c506c_77493.png "屏幕截图")
6. Ingress页面增加域名转发规则信息
   ![输入图片说明](https://foruda.gitee.com/images/1735399689648549556/3d4f8d78_77493.png "屏幕截图")
7. 改进标签显示样式，鼠标悬停展示
   ![输入图片说明](https://foruda.gitee.com/images/1735399387990917764/d06822cb_77493.png "屏幕截图")
8. 优化资源状态样式更小更紧致
   ![输入图片说明](https://foruda.gitee.com/images/1735399419170194492/268b25c8_77493.png "屏幕截图")
9. 丰富Service展示信息
   ![输入图片说明](https://foruda.gitee.com/images/1735399493417833664/fa968343_77493.png "屏幕截图")
10. 突出显示未就绪endpoints
    ![输入图片说明](https://foruda.gitee.com/images/1735399531801079962/9a13cd50_77493.png "屏幕截图")
11. endpoints鼠标悬停展开未就绪IP列表
    ![输入图片说明](https://foruda.gitee.com/images/1735399560648695064/8079b5cf_77493.png "屏幕截图")
12. endpointslice 突出显示未ready的IP及其对应的POD，
    ![输入图片说明](https://foruda.gitee.com/images/1735399614582278222/c1f40aa0_77493.png "屏幕截图")
13. 角色增加延展信息
    ![输入图片说明](https://foruda.gitee.com/images/1735399896080683883/3e9a7359_77493.png "屏幕截图")
14. 角色与主体对应关系
    ![输入图片说明](https://foruda.gitee.com/images/1735399923738735980/c5730152_77493.png "屏幕截图")
15. 界面全量中文化，k8s资源翻译为中文，方便广大用户使用。
    ![输入图片说明](https://foruda.gitee.com/images/1735400283406692980/c778158c_77493.png "屏幕截图")
    ![输入图片说明](https://foruda.gitee.com/images/1735400313832429462/279018dc_77493.png "屏幕截图")
