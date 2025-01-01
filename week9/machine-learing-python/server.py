import datetime
import os
from flask import Flask, request, jsonify
import numpy as np
import pandas as pd
import joblib
import requests

app = Flask(__name__)

# 加载模型和标准化器
model = joblib.load("time_qps_auto_scaling_model.pkl")
scaler = joblib.load("time_qps_auto_scaling_scaler.pkl")

# 从Prometheus 获取QPS
def get_qps_from_prometheus():
    host = os.getenv(
        "PROMETHEUS_HOST",
        "kube-prometheus-stack-prometheus.monitoring.svc.cluster.local:9090"
    )

    url = f"http://{host}/api/v1/query"
    query = 'rate(nginx_ingress_controller_nginx_process_requests_total{service="ingress-nginx-controller-metrics"}[10m])'
    response = requests.get(url,params={"query":query})
    data = response.json()

    qps = float(data["data"]["result"][0]["value"][1])

    print("qps: ", qps)

    return qps

# 定义预测接口
@app.route("/predict", methods=["GET"])
def predict():
    try:
        qps = get_qps_from_prometheus()
        current_time = datetime.datetime.now().strftime("%H:%M:%S")
        # 时间的处理->分钟数

        minutes = (
            pd.to_datetime(current_time, format="%H:%M:%S").hour * 60 
            + pd.to_datetime(current_time, format="%H:%M:%S").minute
        )

        sin_time = np.sin(2 * np.pi * minutes / 1440)
        cos_time = np.cos(2 * np.pi * minutes / 1440)

        # 特征向量
        data = {"QPS": [qps], "sin_time": [sin_time], "cos_time": [cos_time]}

        df = pd.DataFrame(data)
        features_scaled = scaler.transform(df)

        #预测
        prediction = model.predict(features_scaled)

        print("prediction: ", prediction)

        # 为了比卖你实例数过大，限制最大实例是20
        if int(prediction) < 20:
            if int(prediction) == 0:
                return jsonify({"instances": int(1)})
            return jsonify({"instances": int(prediction)})
        else:
            return jsonify({"instances": int(20)})
    except Exception as e:
        return jsonify({"error": str(e)})


# 启动服务
if __name__ == "__main__":
    app.run(host="172.17.99.87", port=8080, debug=True) 
