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