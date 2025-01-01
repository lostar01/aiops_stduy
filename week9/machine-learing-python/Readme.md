#### 使用核心库 Pandas
```bash copy
pip install pandas
```
》主要功能包括数据清洗、数据转换、数据聚合、数据缺失值、数据筛选等
- 常见操作
》read_csv(): 从csv 文件读取数据
》dropna(): 删除缺失值
》groupby(): 数据分组操作
》merge()：合并数据集

#### 使用的核心库 Sklearn
```bash copy
pip install scikit-learn
```
- Skearn
》是一个机器学习的库，提供了大量的算法和工具，用于数据预处理、模型训练、模型评估和模型选择
- 常见功能
》数据预处理： StandardScaler (标准化数据) 、LabelEncoder（标签编码）
》模型训练： 如LinearRegression (线性回归) 、RandomForestClassifier (随机森林分类器)
》模型评估： cross_val_score() (交叉验证) 、 accuracy_score() (准确率评估)

#### 数据载入
```bash copy
df = pd.read_csv("data.csv")
```
- 常见操作
》读取QPS 列： qps_column = df["QPS"]
》读取第二行： second_raw = df.iloc[1]
》读取第二行，QPS 列数据: second_raw_qps = df.loc[1,'QPS']
》筛选QPS > 10 的行： filtered_rows = df[df['QPS'] > 10]

#### 时间格式化的转化
```
df["minutes"] = (
    pd.to_datetime(df["timestamp], format="%H:%M:%S).dt.hour * 60
    + pd.to_datetime(df["timestamp"], format="%H:%M:%S).dt.minute
)
```
》将 timestamp 列转为时间格式，然后转为“从午夜开始的分钟数”， 例如 00:10:00 会被转为10分钟

#### 捕捉时间的周期性
```
df["sin_time"] = np.sin(2 * np.pi * df["minutes] / 1440)
df["cos_time"] = np.cos(2 * np.pi * df["minutes] / 1440)
```

- 使用正弦和余弦函数， 将时间转为周期性的特征
  》1440 为一天的分钟数，通过周期函数让模型更好感知时间如何影响流量



#### 设置特征和预测值
```
x = df[["QPS","sin_time","cos_time"]]
y = df["instances"]
```

- X 表示特征数据，包含：
  》 QPS: 查询每秒的数量
  》 sin_time: 用于捕捉时间周期性信息
  》 cos_time: 用于捕捉时间周期性的信息
- Y 表示目标变量，即我们预测的实力数（instances）


#### 数据分割
```
x_train,x_test, y_train, y_test = train_test_split(x, y, test_size=0.2, random_state=0)
```
- 使用train_test_split 将数据分为训练集和测试集
- test_size=0.2 意味着 20%的数据用于测试， 80%用于训练
- 测试集用来评估模型性能

#### 标准化数据
```
scaler = StandardScaler()
x_train_scaled =  scaler.fit_transform(x_train)
x_test_scaled = scaler.transform(x_test)
```
- StandardScaler 用于将特征数据标准化， 使其均值为0 ，标准差为 1
- 通过fit_transform 对训练集进行标准化，然后用 transform 处理测试集
- 标准化有助于提高模型的性能

#### 模型评估
```
y_pred =  model.coef_ * x_test_scaled + model.intercept_
mse = mean_squared_error(y_test,y_pred.sum(axis=1))
```
- y_pred 是模型对测集的预测结果
- 计算均方差（MSE） 来评估模型的预测性能： MSE 越小，模型越好
- 评估模型可以帮助我们了解其准确性，并决定是否需要改进或调整模型

#### 保存模型
```
joblib.dump(model, "time_qps_auto_scaling_model.pkl")
```
- 使用joblib 库将训练好的模型保存为模型文件
- 直接使用模型文件，而不需要重新训练

#### 保存标准化器
```
joblib.dump(scaler, "time_qps_auto_scaling_scaler.pkl")
```
- 因为对数据进行了标准化，所以需要在推理时保持相同的标准
- 使用joblib 将标准化器保存为文件
- 在使用模型进行预测是，必须确保输入数据经过相同的标准化步骤

#### 使用模型推理
```
#加载模型和标准化器
model = joblib.load("time_qps_auto_scaling_model.pkl")
scaler = joblib.load("time_qps_auto_scaling_scaler.pkl")

# 特征向量
data = {"QPS": [qps], "sin_time": [sin_time], "cos_time": [cos_time]}

df = pd.DataFrame(data)
features_scaled = scaler.transform(df)

# 预测
prediction = model.predict(features_scaled)
```
- 使用 joblib.load 加载保存的模型和标准化器，输入QPS, 时间参数
- 通过调用 predict 方法进行预测， 得到实例数预测结果
- 整合HTTP Server 提供API 推理服务，例如： Python Flask 、Golang Gin 框架等

#### 将模型封装成服务

#### 获取预测结果
- 通过请求推理服务的 /predict API 接口获取推理结果
- 得到预测实例数

#### 开发hpa opertor
```
# init
go mod init github.com/lostar01/hpa-operator
kubebuilder init --domain=aiops.com
kubebuilder create api --group hpa --version v1 --kind PredictHPA
```