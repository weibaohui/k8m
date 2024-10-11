(function () {
    let amisLib = amisRequire('amis');
    let React = amisRequire('react');

    function replacePlaceholders(url, data) {
        // 使用正则表达式匹配 ${} 包裹的变量
        return url.replace(/\$\{([^}]+)\}/g, function (match, key) {
            // key 是 ${} 中的变量名，如 "metadata.namespace"
            // 使用 key 解析 data 对象中的值
            const keys = key.split('.');
            let value = data;
            for (let i = 0; i < keys.length; i++) {
                value = value[keys[i]];
                if (value === undefined) {
                    return match; // 如果没找到对应的值，保留原始字符串
                }
            }
            return value; // 返回找到的值，替换掉 ${} 包裹的变量
        });
    }

    // 构造 URL，附加参数并进行编码
    function appendQueryParam(url, params) {
        const queryString = Object.keys(params)
            .map(key => `${key}=${encodeURIComponent(params[key])}`)
            .join('&');
        // 如果 URL 已经有查询参数，使用 '&' 拼接；否则用 '?' 拼接
        return url.includes('?') ? `${url}&${queryString}` : `${url}?${queryString}`;
    }


    // 自定义组件，props 中可以拿到配置中的所有参数，比如 props.label 是 'Name'
    function SSEComponent(props) {
        // props.url="/k8s/pod/logs/sse/ns/${metadata.namespace}/pod_name/${metadata.name}?a=b&c=d"
        let url = replacePlaceholders(props.url, props.data);
        const params = {
            tailLines: props.data.tailLines,
            sinceTime: props.data.sinceTime,
            follow: props.data.follow,
            previous: props.data.previous,
            timestamps: props.data.timestamps,
            sinceSeconds: props.data.sinceSeconds || ""
        };
        finalUrl = appendQueryParam(url, params);
        let dom = React.useRef(null);
        let eventSourceRef = React.useRef(null);
        const [errorMessage, setErrorMessage] = React.useState('');
        React.useEffect(function () {
            // 处理 SSE 数据
            eventSourceRef.current = new EventSource(finalUrl);

            const handleMessage = (event) => {
                if (dom.current != null) {
                    dom.current.textContent += event.data;
                }
            };

            const handleError = (event) => {
                if (eventSourceRef.current.readyState === EventSource.CLOSED) {
                    setErrorMessage('无日志 连接已关闭');
                } else if (eventSourceRef.current.readyState === EventSource.CONNECTING) {
                    setErrorMessage('无日志 正在尝试重新连接...');
                } else {
                    setErrorMessage('发生未知错误...');
                }
                eventSourceRef.current.close();
            };


            eventSourceRef.current.addEventListener('message', handleMessage);
            eventSourceRef.current.addEventListener('error', handleError);

            return () => {
                if (eventSourceRef.current) {
                    eventSourceRef.current.removeEventListener('message', handleMessage);
                    eventSourceRef.current.removeEventListener('error', handleError);
                    eventSourceRef.current.close();
                }
            };
        }, [url]);
        return React.createElement('div', {
            ref: dom,
            style: {
                whiteSpace: 'pre-wrap'
            }
        }, errorMessage && React.createElement('div', {style: {color: 'red'}}, errorMessage));
    }

    //注册自定义组件，请参考后续对工作原理的介绍
    amisLib.Renderer({
        test: /(^|\/)sse/
    })(SSEComponent);


    function SSEDownloadComponent(props) {
        const [downloading, setDownloading] = React.useState(false);
        let finalUrl = replacePlaceholders(props.url, props.data);
        const params = {
            tailLines: props.data.tailLines,
            sinceTime: props.data.sinceTime,
            previous: props.data.previous,
            timestamps: props.data.timestamps,
            sinceSeconds: props.data.sinceSeconds || ""
        };
        finalUrl = appendQueryParam(finalUrl, params);
        const handleDownload = () => {
            setDownloading(true); // 设置下载状态为 true，显示提示信息

            const anchor = document.createElement('a');
            anchor.href = finalUrl;
            anchor.download = 'log.txt'; // 设置下载的文件名
            document.body.appendChild(anchor);
            anchor.click();
            document.body.removeChild(anchor);

            // 监听下载结束后取消提示
            setTimeout(() => {
                setDownloading(false); // 假设下载立即开始后显示提示的时间
            }, 1000); // 1秒后关闭提示，时间可以根据需要调整
        };

        return React.createElement(
            'div',
            null,
            downloading && React.createElement('p', {
                style: {
                    color: 'red',
                    marginBottom: '10px'
                }
            }, '正在下载，请稍后...'),
            React.createElement(
                'button',
                {
                    onClick: handleDownload,
                    style: {
                        margin: '10px',
                        padding: '8px 16px',
                        backgroundColor: '#4CAF50',
                        color: 'white',
                        border: 'none',
                        cursor: 'pointer',
                        borderRadius: '4px'
                    }
                },
                '下载日志'
            )
        );
    }


    // 注册自定义的 DownloadComponent 组件
    amisLib.Renderer({
        test: /(^|\/)log-download/
    })(SSEDownloadComponent);

    function K8sAgeDisplayComponent(props) {
        // 定义一个状态来存储当前时间，用于计算持续时间
        const [currentTime, setCurrentTime] = React.useState(Date.now());

        // 获取嵌套对象的值的辅助函数
        const getValueByPath = (obj, path) => {
            return path.split('.').reduce((acc, part) => acc && acc[part], obj);
        };

        // 动态获取 creationTimestamp
        const creationTimestamp = getValueByPath(props.data, props.name);

        // 设置定时器每秒更新当前时间
        React.useEffect(() => {
            const interval = setInterval(() => {
                setCurrentTime(Date.now());
            }, 1000); // 每秒更新一次

            // 清除定时器
            return () => clearInterval(interval);
        }, []);

        // 手动更新当前时间，每次 props 变化时触发
        React.useEffect(() => {
            setCurrentTime(Date.now());
        }, [props.data]);

        // 自定义格式化时间的函数
        const formatHumanDuration = (durationInMs) => {
            const seconds = Math.floor(durationInMs / 1000);
            if (seconds < 0) {
                return "0s";
            } else if (seconds < 60 * 2) {
                return `${seconds}s`;
            }
            const minutes = Math.floor(durationInMs / 60000);
            if (minutes < 10) {
                const s = Math.floor((durationInMs / 1000) % 60);
                return s === 0 ? `${minutes}m` : `${minutes}m${s}s`;
            } else if (minutes < 60 * 3) {
                return `${minutes}m`;
            }
            const hours = Math.floor(durationInMs / (60 * 60 * 1000));
            if (hours < 8) {
                const m = Math.floor((durationInMs / 60000) % 60);
                return m === 0 ? `${hours}h` : `${hours}h${m}m`;
            } else if (hours < 48) {
                return `${hours}h`;
            }
            const days = Math.floor(hours / 24);
            if (days < 8) {
                const h = hours % 24;
                return h === 0 ? `${days}d` : `${days}d${h}h`;
            } else if (days < 365 * 2) {
                return `${days}d`;
            }
            const years = Math.floor(days / 365);
            const dy = days % 365;
            return years < 8 ? (dy === 0 ? `${years}y` : `${years}y${dy}d`) : `${Math.floor(days / 365)}y`;
        };

        // 如果 creationTimestamp 无效，返回 N/A
        if (!creationTimestamp) {
            return React.createElement('span', null, 'N/A');
        }

        // 计算时间差
        const durationInMs = currentTime - new Date(creationTimestamp).getTime();
        const formattedTime = formatHumanDuration(durationInMs);

        // 显示格式化的时间
        return React.createElement('span', null, formattedTime);
    }
    // 注册自定义组件
    amisLib.Renderer({
        test: /(^|\/)k8sAge/
    })(K8sAgeDisplayComponent);


    function K8sPodReadyComponent(props) {
        // 获取Pod状态中的容器状态列表
        const containerStatuses = props.data.status?.containerStatuses || [];

        // 从 Pod 的 spec 中获取定义的容器数量
        const containerSpecs = props.data.spec?.containers || [];

        // 计算处于 Ready 状态的容器数量
        const readyCount = containerStatuses.filter(status => status.ready).length;

        // 总的容器数量应该从 spec 中获取
        const totalCount = containerSpecs.length;

        // 将 Ready 数量与总数量格式化为 N/M 形式
        const readyStatus = `${readyCount}/${totalCount}`;

        return React.createElement('span', null, readyStatus);
    }

// 注册自定义组件
    amisLib.Renderer({
        test: /(^|\/)k8sPodReady/
    })(K8sPodReadyComponent);

    function NodeRolesComponent(props) {
        const labels = props.data.metadata?.labels || {};
        const roles = Object.keys(labels).filter(label => label.startsWith('node-role.kubernetes.io/'));

        // 角色汉字映射
        const roleMap = {
            'master': '主节点',
            'control-plane': '控制平面',
            'worker': '工作节点',
            'ingress': '入口节点',
            'storage': '存储节点',
            'compute': '计算节点',
        };

        const displayedRoles = roles.map(role => {
            const roleKey = role.replace('node-role.kubernetes.io/', '');
            return roleMap[roleKey] || roleKey; // 如果没有映射，则显示原角色
        });

        return React.createElement('span', null,
            displayedRoles.length > 0
                ? displayedRoles.join(', ')
                : ''
        );
    }

// 注册自定义组件
    amisLib.Renderer({
        test: /(^|\/)nodeRoles/
    })(NodeRolesComponent);


    function K8sLabelsComponent(props) {
        const labels = props.data.metadata?.labels || {};

        // 将 labels 对象转换为数组并生成每个标签的元素
        const labelElements = Object.keys(labels).map((key, index) => {
            return React.createElement('div', {key: index, style: {marginTop: '5px'}},
                React.createElement('span', {
                    key: index,
                    className: 'label bg-gray-100 text-black',
                }, `${key}: ${labels[key]}`)
            );

        });

        return React.createElement('div', null, labelElements.length > 0 ? labelElements : '');
    }

// 注册自定义组件
    amisLib.Renderer({
        test: /(^|\/)k8sLabels/
    })(K8sLabelsComponent);


    function K8sConditionsComponent(props) {
        const conditions = props.data.status?.conditions || [];

        const conditionElements = conditions.map((condition, index) => {
            const tagClass = condition.status === 'True' || (condition.status === 'False' &&( condition.type.includes('Pressure')||condition.type.includes("Unavailable")))
                ? 'label label-success'  // 正常情况
                : 'label label-danger';   // 问题情况

            return React.createElement('div', {key: index, style: {marginTop: '5px'}},
                React.createElement('span', {
                    key: index,
                    className: tagClass,
                }, condition.type)
            );
        });

        return React.createElement('div', null, conditionElements);
    }


// 注册自定义组件
    amisLib.Renderer({
        test: /(^|\/)k8sConditions/
    })(K8sConditionsComponent);


    function CIExecutionAgeDisplayComponent(props) {
        // 定义一个状态来存储当前时间，用于计算持续时间
        const [currentTime, setCurrentTime] = React.useState(Date.now());

        // 获取嵌套对象的值的辅助函数
        function getValueByPath(obj, path) {
            return path.split('.').reduce(function (acc, part) {
                return acc && acc[part];
            }, obj);
        }

        // 动态获取 build_duration, build_start_time, build_end_time
        const buildDuration = getValueByPath(props.data, 'build_duration'); // 秒数
        const buildStartTime = getValueByPath(props.data, 'build_start_time'); // 时间字符串
        const buildEndTime = getValueByPath(props.data, 'build_end_time'); // 时间字符串
        // 设置定时器每秒更新当前时间
        React.useEffect(function () {
            const interval = setInterval(function () {
                setCurrentTime(Date.now());
            }, 1000); // 每秒更新一次

            // 清除定时器
            return function () {
                clearInterval(interval);
            };
        }, []);

        // 手动更新当前时间，每次 props 变化时触发
        React.useEffect(function () {
            setCurrentTime(Date.now());
        }, [props.data]);

        // 自定义格式化时间的函数
        function formatHumanDuration(durationInMs) {
            const seconds = Math.floor(durationInMs / 1000);
            if (seconds < 0) {
                return "0s";
            } else if (seconds < 60 * 2) {
                return seconds + "s";
            }
            const minutes = Math.floor(durationInMs / 60000);
            if (minutes < 10) {
                const s = Math.floor((durationInMs / 1000) % 60);
                return s === 0 ? minutes + "m" : minutes + "m" + s + "s";
            } else if (minutes < 60 * 3) {
                return minutes + "m";
            }
            const hours = Math.floor(durationInMs / (60 * 60 * 1000));
            if (hours < 8) {
                const m = Math.floor((durationInMs / 60000) % 60);
                return m === 0 ? hours + "h" : hours + "h" + m + "m";
            } else if (hours < 48) {
                return hours + "h";
            }
            const days = Math.floor(hours / 24);
            if (days < 8) {
                const h = hours % 24;
                return h === 0 ? days + "d" : days + "d" + h + "h";
            } else if (days < 365 * 2) {
                return days + "d";
            }
            const years = Math.floor(days / 365);
            const dy = days % 365;
            return years < 8 ? (dy === 0 ? years + "y" : years + "y" + dy + "d") : Math.floor(days / 365) + "y";
        }

        let formattedTime = "-"; // 默认显示为 0s

        if (buildDuration !== undefined && buildDuration !== null && buildDuration!=='0') {
            // 如果 build_duration 有值，直接显示秒数
            formattedTime = formatHumanDuration(buildDuration*1000);
        } else if (buildStartTime) {
            // 如果没有 build_duration，但有 build_start_time
            const startTime = new Date(buildStartTime).getTime();
            const endTime = buildEndTime ? new Date(buildEndTime).getTime() : currentTime;
            const durationInMs = endTime - startTime;
            formattedTime = formatHumanDuration(durationInMs);
        }

        // 显示格式化的时间
        return React.createElement('span', null, formattedTime);
    }
// 注册自定义组件
    amisLib.Renderer({
        test: /(^|\/)ciExecutionAge/
    })(CIExecutionAgeDisplayComponent);

    // 定义自动转换内存单位的过滤器
    amisLib.registerFilter('autoConvertMemory', function (input) {
        // 定义单位和对应的倍数
        const unitMultipliers = {
            Ki: 1024,
            Mi: 1024 ** 2,
            Gi: 1024 ** 3,
            Ti: 1024 ** 4,
            Pi: 1024 ** 5,
            Ei: 1024 ** 6,
        };

        // 使用正则表达式匹配数值和单位
        const match = input.match(/(\d+)([KMGTPE]i)/i);

        if (!match) {
            return input; // 无法匹配时，返回原始输入
        }

        const value = parseInt(match[1], 10);
        const unit = match[2];

        // 将值转换为字节数
        const bytes = value * unitMultipliers[unit];

        // 自动选择适合的单位显示
        if (bytes < 1024 ** 2) {
            return `${(bytes / 1024).toFixed(1)}Ki`;
        } else if (bytes < 500 * 1024 ** 2) {
            return `${(bytes / 1024 ** 2).toFixed(1)}Mi`;
        } else {
            return `${(bytes / 1024 ** 3).toFixed(1)}Gi`;
        }
    });
    // 最简化展示镜像名称及tag
    amisLib.registerFilter('simpleImageName', function (input) {
        // 如果输入不是字符串，直接返回原始输入
        if (typeof input !== 'string') {
            return input;
        }

        // 分割镜像名称，移除注册表地址部分
        const parts = input.split('/');
        const imageName = parts[parts.length - 1]; // 获取最后一部分

        // 去除镜像版本
        const [name, tag] = imageName.split(':');

        // 返回基本的名称和 tag，tag 默认为 "latest"
        return `${name}:${tag || 'latest'}`;
    });
})();