import sys
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

from threading import Thread
import socket
import traceback
from datetime import datetime
import csv
import time

#HOST = '10.0.0.253'
HOST = "192.168.43.98"
PORT = 8081

addr = (HOST, PORT)

client_sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
client_sock.connect(addr)

now = datetime.now().strftime("%Y%m%d-%H%M%S")
FILENAME = str(now) + ".csv"
print(FILENAME)
log_file = open(FILENAME, "w")

parameters = dict.fromkeys(['time',
                            'raw_cell_0', 'raw_cell_1', 'raw_cell_2', 'raw_cell_3',
                            'temperature', 'pressure', 'hymidity',
                            'flow_air', 'flow_test_gas'
                            'amp2ppm_0', 'amp2ppm_1', 'amp2ppm_2', 'amp2ppm_3',
                            'base_line_0', 'base_line_1', 'base_line_2', 'base_line_3'])

columns = ['time',
           'raw_cell_0', 'raw_cell_1', 'raw_cell_2', 'raw_cell_3',
           'temperature', 'pressure', 'hymidity',
           'flow_air', 'flow_test_gas'
           'amp2ppm_0', 'amp2ppm_1', 'amp2ppm_2', 'amp2ppm_3',
           'base_line_0', 'base_line_1', 'base_line_2', 'base_line_3']


@QtCore.pyqtSlot()
def run1():
    print("Job 1")

    data = client_sock.recv(1024)
    server_data = data.decode(encoding="utf-8")
    server_data = server_data[:-1]
    print("server_data", server_data)
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
    def __init__(self):
        super(GasBench, self).__init__()
        self.gui = Ui_MainWindow()
        self.gui.setupUi(self)

        timer = QtCore.QTimer(self)
        timer.timeout.connect(self.runner)
        timer.start(1000)

        timer2 = QtCore.QTimer(self)
        timer2.timeout.connect(self.ping)
        timer2.start(10000)

        timer3 = QtCore.QTimer(self)
        timer3.timeout.connect(self.get_flow)
        timer3.start(1000)

        timer4 = QtCore.QTimer(self)
        timer4.timeout.connect(self.get_ppm)
        timer4.start(30000)

        self.init_gui()

    def init_gui(self):
        # self.gui.comboBox.addItems(port_scanner())

        self.gui.SendFlowParams.clicked.connect(
            self.set_flow_controller_params)
        self.gui.GetGASettings.clicked.connect(self.get_gas_sensor_params)
        self.gui.SetGASettings.clicked.connect(self.send_gas_sensor_params)
        self.gui.pushButton.clicked.connect(self.send_flow_value)

        self.curve = []
        self.timer_graph = [[], [], [], []]
        self.data_arrays = [np.array([]), np.array(
            []), np.array([]), np.array([])]

        self.curve.append(self.gui.graph_1.plot(pen="y"))
        self.curve.append(self.gui.graph_2.plot(pen="y"))
        self.curve.append(self.gui.graph_3.plot(pen="y"))
        self.curve.append(self.gui.graph_4.plot(pen="y"))

        self.i = 0

        self.gui.reset_graph_0.clicked.connect(lambda: self.graph_reset(0))
        self.gui.reset_graph_1.clicked.connect(lambda: self.graph_reset(1))
        self.gui.reset_graph_2.clicked.connect(lambda: self.graph_reset(2))
        self.gui.reset_graph_3.clicked.connect(lambda: self.graph_reset(3))

        self.gas_types_widget = [
            self.gui.gas_type_0, self.gui.gas_type_1, self.gui.gas_type_2, self.gui.gas_type_3]

        self.show()

    def print_output(self, result):
        # print(result)
        #self.send_to_server("get_raw_data", "")
        self.server_receive(result)

    def thread_complete(self):
        print("complete")

    def runner(self):
        thread_pool = QtCore.QThreadPool.globalInstance()
        #print("server check")
        worker = Worker(run1)
        worker.signals.result.connect(self.print_output)
        worker.signals.finished.connect(self.thread_complete)
        thread_pool.start(worker)

    def ping(self):
        self.send_to_server("get_raw_data", "")

    def get_flow(self):
        self.send_to_server("get_flow", "")

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

    def send_gas_sensor_params(self):
        settings = []
        send_string = ""

        settings.append(self.humidity_2.text())

        for i in range(0, 4):
            for j in range(0, 11):
                settings.append(self.gui.tableGASettings.item(j, i).text())

        for i in range(0, 4):
            number_of_gas = 0
            for j in range(0, len(self.gas) - 1):
                if self.gas[j] == settings[i * 11]:
                    number_of_gas = j
            settings[i * 11] = str(number_of_gas)

        for i in settings:
            send_string = send_string + str(i) + " "
        send_string = send_string + "|"

        print("settings >>>>", send_string)

        # value_0 = self.gui.tableGASettings.item(10, 0).text()
        # value_1 = self.gui.tableGASettings.item(10, 1).text()
        # value_2 = self.gui.tableGASettings.item(10, 2).text()
        # value_3 = self.gui.tableGASettings.item(10, 3).text()

        # value_4 = self.gui.tableGASettings.item(8, 0).text()
        # value_5 = self.gui.tableGASettings.item(8, 1).text()
        # value_6 = self.gui.tableGASettings.item(8, 2).text()
        # value_7 = self.gui.tableGASettings.item(8, 3).text()

        # parameters['amp2ppm_0'] = value_4
        # parameters['amp2ppm_1'] = value_5
        # parameters['amp2ppm_2'] = value_6
        # parameters['amp2ppm_3'] = value_7

        # parameters['base_line_0'] = value_0
        # parameters['base_line_1'] = value_1
        # parameters['base_line_2'] = value_2
        # parameters['base_line_3'] = value_3

        # send_string = str(value_0) + " " + str(value_1) + \
        #     " " + str(value_2) + " " + str(value_3) + " " + str(value_4) + " " + \
        #     str(value_5) + " " + str(value_6) + " " + str(value_7) + " " + "|"
        # self.send_to_server("set_ga", send_string)

        # with open(FILENAME, "w", newline="") as file:
        #     columns = columns
        #     writer = csv.DictWriter(file, fieldnames=columns)
        #     writer.writeheader()

        # writer.writerow(parameters)
        # time.sleep(2)

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

    i = 0
    x = []

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

            print(server_array)
            self.graphs_update(server_array[1:])

            # wrong name of the textfield!!!!!
            self.sampling_depth = self.gui.GasType_3.text()
            for i in self.data_arrays:
                self.average = np.mean(i[-self.sampling_depth:])
            self.gui.average_0.setText

        elif server_array[0] == "ga_options":
            print(server_array)

            # wrong widjet name!!!!!!
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

        elif server_array[0] == "get_flow":
            self.gui.tableFlowSettings.setItem(
                1, 0, QTableWidgetItem(server_array[1]))
            self.gui.tableFlowSettings.setItem(
                3, 0, QTableWidgetItem(server_array[2]))
        elif server_array[0] == "get_ppm":
            self.gui.temperature.setText(server_array[1])
            self.gui.humidity.setText(server_array[2])
            self.gui.pressure.setText(server_array[3])

            self.gui.ppm_0.setText(server_array[5])
            self.gui.ppm_1.setText(server_array[7])
            self.gui.ppm_2.setText(server_array[9])
            self.gui.ppm_3.setText(server_array[11])

    def graphs_update(self, server_array):
        #curve = []

        print("graph update: ", server_array)

        current_time = datetime.now()

        current_time_representation = current_time.hour + \
            current_time.minute / 60 + current_time.second / 3600
        for i in self.timer_graph:
            i.append(current_time_representation)

        # self.curve.append(self.gui.graph_1.plot(pen="y"))
        # self.curve.append(self.gui.graph_2.plot(pen="y"))
        # self.curve.append(self.gui.graph_3.plot(pen="y"))
        # self.curve.append(self.gui.graph_4.plot(pen="y"))

        for c in self.curve:
            c.clear()

        x_in = []
        for j in server_array:
            x_in.append(np.array(j, dtype=float))

        for k in range(0, len(self.data_arrays)):
            self.data_arrays[k] = np.append(self.data_arrays[k], x_in[k])
            forPainting = np.column_stack(
                (self.timer_graph[k], self.data_arrays[k]))

            self.curve[k].setData(forPainting[:])

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


def port_scanner():
    ports = QSerialPortInfo().availablePorts()
    port_name = []
    for port in ports:
        port_name.append(port.portName())

    return port_name


if __name__ == "__main__":
    app = QtWidgets.QApplication([])
    application = GasBench()
    application.show()

    sys.exit(app.exec())
