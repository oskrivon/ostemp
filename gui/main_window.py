from cProfile import run
import sys
from telnetlib import GA
from unicodedata import name
from unittest import case
from PyQt5 import QtCore, QtWidgets
from PyQt5.QtGui import QIcon
from pyqtgraph.graphicsItems.GridItem import GridItem
from gui import Ui_MainWindow
from PyQt5.QtSerialPort import QSerialPort, QSerialPortInfo
from pyqtgraph.Qt import QtGui, QtCore
from PyQt5.QtCore import QThreadPool, QRunnable, QObject, pyqtSignal
from PyQt5.QtWidgets import QTableWidgetItem

import numpy as np
import pyqtgraph as pg

import threading
import socket
import traceback
from datetime import datetime
import csv
import time
import yaml

with open("gui\dist\config.yaml") as f:
    config = yaml.safe_load(f)

print(config)
print("ip: ", config["netConfig"]["ip"])

HOST = config["netConfig"]["ip"]
PORT = config["netConfig"]["port"]

addr = (HOST, PORT)

new_server_data = ""

client_sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
client_sock.connect(addr)

now = datetime.now().strftime("%Y%m%d-%H%M%S")
columns = ['time',
           'raw_cell_0', 'raw_cell_1', 'raw_cell_2', 'raw_cell_3',
           'temperature', 'pressure', 'hymidity',
           'flow_air', 'flow_test_gas',
           'amp2ppm_0', 'amp2ppm_1', 'amp2ppm_2', 'amp2ppm_3',
           'base_line_0', 'base_line_1', 'base_line_2', 'base_line_3']

FILENAME = str(now) + ".csv"
with open(FILENAME, 'w') as log_file:
    writer = csv.DictWriter(log_file, fieldnames=columns)
    writer.writeheader()

print(FILENAME)

server_data = ""


class MyThread(QtCore.QThread):
    about_new_log = QtCore.pyqtSignal()

    def run(self):
        while 1:
            data_raw = client_sock.recv(1024)
            data_clean = data_raw.decode(encoding="utf-8")
            data_clean = data_clean[:-1]

            global server_data
            server_data = data_clean
            #some_text = some_text + "."
            self.about_new_log.emit()

            # sleep(1)


@QtCore.pyqtSlot()
def ping():
    data = b''

    while True:
        data_raw = client_sock.recv(1)
        if data_raw == b'|':
            break
        data = data + data_raw
        #print(">>> data: ", data)
    # print(">>> raw data from server", type(
       # data), data_raw, " len:", len(data_raw))
    server_data = data.decode(encoding="utf-8")
    server_data = server_data[:-1]
    print(">>> server_data", server_data)
    return server_data


class WorkerSignals(QObject):
    finished = QtCore.pyqtSignal()
    error = QtCore.pyqtSignal(tuple)
    result = QtCore.pyqtSignal(object)
    progress = QtCore.pyqtSignal(int)


class Worker(QtCore.QRunnable):
    def __init__(self, fn, *args, **kwargs):
        super(Worker, self).__init__()
        self.fn = fn
        self.args = args
        self.kwargs = kwargs
        self.signals = WorkerSignals()

    def run(self):
        try:
            result = self.fn(*self.args, **self.kwargs)
        except:
            traceback.print_exc()
            exctype, value = sys.exc_info()[:2]
            self.signals.error.emit((exctype, value, traceback.format_exc()))
        else:
            # Return the result of the processing
            self.signals.result.emit(result)
        finally:
            self.signals.finished.emit()  # Done


class GasBench(QtWidgets.QMainWindow):
    my_signal = QtCore.pyqtSignal(list, name='my_signal')

    def __init__(self):
        super(GasBench, self).__init__()
        self.gui = Ui_MainWindow()
        self.gui.setupUi(self)

        #timer = QtCore.QTimer(self)
        # timer.timeout.connect(self.runner)
        # timer.start(1000)

        timer2 = QtCore.QTimer(self)
        timer2.timeout.connect(self.ping)
        timer2.start(10000)

        timer3 = QtCore.QTimer(self)
        timer3.timeout.connect(self.get_flow)
        timer3.start(1000)

        timer4 = QtCore.QTimer(self)
        timer4.timeout.connect(self.get_ppm)
        timer4.start(30000)

        self.system_state = dict.fromkeys(columns)

        timer5 = QtCore.QTimer(self)
        timer5.timeout.connect(self.log_update)
        timer5.start(10000)

        self.my_signal.connect(self.raw_printing, QtCore.Qt.QueuedConnection)

        self.init_gui()

    def init_gui(self):
        self.gui.SendFlowParams.clicked.connect(
            self.set_flow_controller_params)
        self.gui.GetGASettings.clicked.connect(self.get_gas_sensor_params)
        self.gui.SetGASettings.clicked.connect(self.send_gas_sensor_params)
        self.gui.pushButton.clicked.connect(self.send_flow_value)
        self.gui.average_calc.clicked.connect(self.average_calculation)
        self.gui.SetGasType.clicked.connect(self.set_gas_type)
        self.gui.CleanAir.clicked.connect(self.clean_air)

        self.curve = []
        self.timer_graph = [[], [], [], []]
        self.data_arrays = [np.array([]), np.array(
            []), np.array([]), np.array([])]

        self.flow_array_1 = []
        self.flow_array_2 = []

        self.curve.append(self.gui.graph_1.plot(pen="y"))
        self.curve.append(self.gui.graph_2.plot(pen="y"))
        self.curve.append(self.gui.graph_3.plot(pen="y"))
        self.curve.append(self.gui.graph_4.plot(pen="y"))

        self.gui.reset_graph_0.clicked.connect(lambda: self.graph_reset(0))
        self.gui.reset_graph_1.clicked.connect(lambda: self.graph_reset(1))
        self.gui.reset_graph_2.clicked.connect(lambda: self.graph_reset(2))
        self.gui.reset_graph_3.clicked.connect(lambda: self.graph_reset(3))

        self.gas_types_widget = [
            self.gui.gas_type_0, self.gui.gas_type_1, self.gui.gas_type_2, self.gui.gas_type_3]

        self.average_widjets = [
            self.gui.average_0, self.gui.average_1, self.gui.average_2, self.gui.average_3]

        self.thread = MyThread()
        self.thread.about_new_log.connect(self.raw_printing)
        self.thread.start()

        self.show()

    def raw_printing(self):
        global server_data
        print(">>>>> server data: ", server_data)
        self.server_receive(server_data)

    def print_output(self, result):
        self.server_receive(result)

    def thread_complete(self):
        print("complete")

    def runner(self):
        thread_pool = QtCore.QThreadPool.globalInstance()
        worker = Worker(ping)
        worker.signals.result.connect(self.print_output)
        worker.signals.finished.connect(self.thread_complete)
        thread_pool.start(worker)

    def ping(self):
        self.send_to_server("get_raw_data", "")

    def get_flow(self):
        self.send_to_server("get_flow", "")

    def clean_air(self):
        self.send_to_server("clean_air", "")

    def get_ppm(self):
        self.send_to_server("get_ppm", "")

    def send_gas_controller_params(self):
        gas_type = self.gui.GasType.text()
        target_flow = self.gui.TargetFlow.text()
        target_concentration = self.gui.TargetConc.text()
        send_string = (
            str(gas_type)
            + " "
            + str(target_flow)
            + " "
            + str(target_concentration)
            + "|"
        )

        self.send_to_server("set_flow_set", send_string)
        print(gas_type, target_flow, target_concentration)

    def set_gas_type(self):
        column = int(self.gui.comboBoxCell.currentText())
        gasTemplate = self.gui.comboBoxGasType.currentText()
        gasType = config["gases"][gasTemplate]

        self.customFill(column, 0, gasTemplate)
        self.customFill(column, 1, gasType["v_ref"])
        self.customFill(column, 2, gasType["v_ref_comp"])
        self.customFill(column, 3, gasType["afe_bias"])
        self.customFill(column, 4, gasType["afe_r_gain"])
        self.customFill(column, 5, gasType["rangeMin"])
        self.customFill(column, 6, gasType["rangeMax"])
        self.customFill(column, 7, gasType["resolution"])
        self.customFill(column, 8, gasType["amp2ppm"])
        self.customFill(column, 9, gasType["responseTime"])
        self.customFill(column, 10, gasType["baselineShift"])

    def customFill(self, column, position, value):
        self.gui.tableGASettings.setItem(
            position, column, QTableWidgetItem(str(value)))

    def send_gas_sensor_params(self):
        settings = []
        send_string = ""

        for i in range(0, 4):
            for j in range(0, 11):
                settings.append(self.gui.tableGASettings.item(j, i).text())

        for i in range(0, 4):
            number_of_gas = 0
            for j in range(0, len(self.gas) - 1):
                if self.gas[j] == settings[i * 11]:
                    number_of_gas = j
            settings[i * 11] = str(number_of_gas)

        xxx = [self.gui.humidity_2.text().encode('utf-8')]

        xxx.extend(settings)
        settings = xxx
        print("___________________ip: ", xxx)

        for i in settings:
            send_string = send_string + str(i) + " "
        send_string = send_string + "|"

        print("settings >>>>", settings)

        self.send_to_server("set_ga", send_string)
        self.get_gas_sensor_params()

    def send_flow_value(self):
        value_0 = self.gui.tableFlowSettings.item(0, 0).text()
        value_1 = self.gui.tableFlowSettings.item(2, 0).text()

        self.gui.TargetFlow.setText(str(float(value_0) + float(value_1)))
        self.gui.TargetConc.setText(
            str(float(value_1) / (float(value_1) + float(value_0))))

        send_string = str(value_0) + " " + str(value_1) + "|"
        self.send_to_server("set_flow", send_string)

    def get_gas_sensor_params(self):
        self.send_to_server("get_ga", "")

    def set_flow_controller_params(self):
        flow = float(self.gui.TargetFlow.text())
        conc = float(self.gui.TargetConc.text())
        if conc >= 1:
            print("error conc!!!")
        else:
            flow_gas = flow * conc
            flow_air = flow - flow_gas

            self.gui.tableFlowSettings.setItem(
                0, 0, QTableWidgetItem(str(flow_air)))
            self.gui.tableFlowSettings.setItem(
                2, 0, QTableWidgetItem(str(flow_gas)))

            send_string = str(flow_air) + " " + str(flow_gas) + "|"
            self.send_to_server("set_flow", send_string)

        self.send_to_server("get_flow", "")

    def send_to_server(self, command_handler, payload):
        if payload == "":
            send_string = command_handler + "|"
        else:
            send_string = command_handler + " " + payload + "|"

        client_sock.sendall(bytes(send_string, "UTF-8"))

    gas = dict()

    gas[0] = "non"
    gas[1] = "CO"
    gas[2] = "NO2"
    gas[3] = "SO2"
    gas[4] = "O3"
    gas[5] = "CH2O"
    gas[6] = "H2S"
    gas[7] = "NO"
    gas[8] = "HCl"
    gas[9] = "NH3"
    gas[10] = "CO2"
    gas[11] = "CH4"
    gas[12] = "HF"
    gas[12] = "Cl2"

    time = []

    def server_receive(self, server_data):
        server_array = server_data.split(" ")

        if server_array[0] == "raw_data":
            self.gui.current_0.setText(server_array[1])
            self.gui.current_1.setText(server_array[2])
            self.gui.current_2.setText(server_array[3])
            self.gui.current_3.setText(server_array[4])

            self.system_state["raw_cell_0"] = float(server_array[1])
            self.system_state["raw_cell_1"] = float(server_array[2])
            self.system_state["raw_cell_2"] = float(server_array[3])
            self.system_state["raw_cell_3"] = float(server_array[4])

            print(server_array)
            self.graphs_update(server_array[1:])

        elif server_array[0] == "ga_options":
            print(server_array)

            # wrong widjet name!!!!!!
            #self.gui.humidity_2.setText(
            #   bytes.fromhex(server_array[2]).decode('utf-8'))
            self.gui.humidity_2.setText(server_array[2])

            sells_settings = server_array[3:]
            print(">>>>>>", sells_settings)

            for i in range(0, 11):
                for j in range(0, 4):
                    self.gui.tableGASettings.setItem(
                        i, j, QTableWidgetItem(sells_settings[i + j * 11]))

            for i in range(0, 4):
                self.gui.tableGASettings.setItem(0, i, QTableWidgetItem(
                    self.gas[int(sells_settings[i * 11])]))

            for i in range(0, 4):
                self.gas_types_widget[i].setText(
                    self.gas[int(sells_settings[i * 11])])

            self.system_state["amp2ppm_0"] = float(sells_settings[8])
            self.system_state["amp2ppm_1"] = float(sells_settings[19])
            self.system_state["amp2ppm_2"] = float(sells_settings[30])
            self.system_state["amp2ppm_3"] = float(sells_settings[41])

            self.system_state["base_line_0"] = float(sells_settings[10])
            self.system_state["base_line_1"] = float(sells_settings[21])
            self.system_state["base_line_2"] = float(sells_settings[32])
            self.system_state["base_line_3"] = float(sells_settings[43])

        elif server_array[0] == "get_flow":
            flow_air = 0.0
            flow_test_gas = 0.0

            if server_array[1] == "":
                flow_air = 0
            else:
                flow_air = float(server_array[1])

            if server_array[2] == "":
                flow_test_gas = 0
            else:
                flow_test_gas = float(server_array[2])

            self.gui.tableFlowSettings.setItem(
                1, 0, QTableWidgetItem(str(flow_air)))
            self.gui.tableFlowSettings.setItem(
                3, 0, QTableWidgetItem(str(flow_test_gas)))

            self.system_state["flow_air"] = flow_air
            self.system_state["flow_test_gas"] = flow_test_gas

            self.flow_graph_update([flow_air, flow_test_gas])
        elif server_array[0] == "get_ppm":
            self.gui.temperature.setText(server_array[1])
            self.gui.humidity.setText(server_array[2])
            self.gui.pressure.setText(server_array[3])

            self.system_state["temperature"] = float(server_array[1])
            self.system_state["pressure"] = float(server_array[2])
            self.system_state["hymidity"] = float(server_array[3])

            self.gui.ppm_0.setText(server_array[5])
            self.gui.ppm_1.setText(server_array[7])
            self.gui.ppm_2.setText(server_array[9])
            self.gui.ppm_3.setText(server_array[11])

        elif server_array[0] == "busy":
            #self.gui.GetGASettings.setStyleSheet('background: rgb(255,0,0);')
            self.gui.GetGASettings.setEnabled(False)
            self.gui.SetGASettings.setEnabled(False)
            # self.gui.temperature.setText("axaxaxaxaxa")

        elif server_array[0] == "free":
            #self.gui.GetGASettings.setStyleSheet('background: rgb(0,0,0);')
            self.gui.GetGASettings.setEnabled(True)
            self.gui.SetGASettings.setEnabled(True)
            # self.gui.temperature.setText("-----------")

    def average_calculation(self):
        sampling_depth = int(self.gui.GasType_3.text())
        averages_array = []
        if sampling_depth >= 1:
            for i in self.data_arrays:
                average = np.mean(i[-sampling_depth:])
                averages_array.append(average)
            #average = np.mean(self.data_arrays[0][-sampling_depth:])
        else:
            average = 0

        for i in range(0, 4):
            self.average_widjets[i].setText(str(averages_array[i]))

    def graphs_update(self, server_array):
        current_time = datetime.now()

        current_time_representation = current_time.hour + \
            current_time.minute / 60 + current_time.second / 3600
        for i in self.timer_graph:
            i.append(current_time_representation)

        for c in self.curve:
            c.clear()

        x_in = []
        for i in range(0, 4):
            x_in.append(np.array(server_array[i], dtype=float))

        for k in range(0, len(self.data_arrays)):
            self.data_arrays[k] = np.append(self.data_arrays[k], x_in[k])
            forPainting = np.column_stack(
                (self.timer_graph[k], self.data_arrays[k]))

            self.curve[k].setData(forPainting[:])

    def flow_graph_update(self, flow_data):
        x_in_1 = []
        x_in_2 = []

        #self.flow_curve = self.gui.graph_5.plot(pen="y")

        x_in_1.append(np.array(flow_data[0], dtype=float))
        x_in_2.append(np.array(flow_data[1], dtype=float))

        self.flow_array_1 = np.append(self.flow_array_1, x_in_1)
        self.flow_array_2 = np.append(self.flow_array_2, x_in_2)

        self.gui.graph_5.plot(self.flow_array_1[:], pen="y")
        self.gui.graph_5.plot(self.flow_array_2[:], pen="r")

        #self.flow_curve.setData(self.flow_array_1[:], self.flow_array_2[:])
        # self.flow_curve.setData(self.flow_array_2[:])

    def graph_reset(self, button_id):
        try:
            self.curve[button_id].clear()
            self.data_arrays[button_id] = []
            self.timer_graph[button_id] = []

            forPainting = np.column_stack(
                (self.timer_graph[button_id], self.data_arrays[button_id]))

            self.curve[button_id].setData(forPainting[:])

        except Exception:
            print("zeroing error")
        else:
            print("zeroing ok")

    def log_update(self):
        now = datetime.now().strftime("%Y%m%d-%H%M%S")
        print(self.system_state)
        self.system_state["time"] = now

        with open(FILENAME, 'a') as log_file:
            writer = csv.DictWriter(log_file, fieldnames=columns)
            writer.writerow(self.system_state)

        print("---------csv ok")


def port_scanner():
    ports = QSerialPortInfo().availablePorts()
    port_name = []
    for port in ports:
        port_name.append(port.portName())

    return port_name


def read_server():
    while 1:
        data = client_sock.recv(1024)
        new_server_data = data.decode(encoding="utf-8")
        new_server_data = new_server_data[:-1]
        print(">>>>> new data:", new_server_data)
        application.server_receive(new_server_data)


if __name__ == "__main__":
    app = QtWidgets.QApplication([])
    application = GasBench()

    #server_thread = threading.Thread(target=read_server)
    # server_thread.start()
    application.show()

    sys.exit(app.exec())
