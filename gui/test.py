from PyQt5 import QtCore, QtGui, QtWidgets

import time
import traceback, sys

uart_result = ['1','2', '3', '4', '5', '6']

#Running all these methods in parallel
@QtCore.pyqtSlot()
def run1():
    print("Job 1")
    return uart_result

@QtCore.pyqtSlot()
def run2():
    print("Job 2")
    return uart_result

@QtCore.pyqtSlot()
def run3():
    print("Job 3")
    return uart_result

@QtCore.pyqtSlot()
def run4():
    print("Job 4")
    return uart_result

class WorkerSignals(QtCore.QObject):
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
            self.signals.result.emit(result)  # Return the result of the processing
        finally:
            self.signals.finished.emit()  # Done



class MainWindow(QtWidgets.QMainWindow):
    def __init__(self, *args, **kwargs):
        super(MainWindow, self).__init__(*args, **kwargs)
        b = QtWidgets.QPushButton("START!")
        b.pressed.connect(self.runner)

        w = QtWidgets.QWidget()
        layout = QtWidgets.QVBoxLayout(w)
        layout.addWidget(b)
        self.setCentralWidget(w)

    def print_output(self, uart_list):
        print(uart_list)

    def thread_complete(self):
        print("THREAD COMPLETE!")

    def runner(self):
        thread_pool = QtCore.QThreadPool.globalInstance()
        print("Multithreading with maximum %d threads" % thread_pool.maxThreadCount())
        print("You pressed the Test button")
        for task in (run1, run2, run3, run4):
            worker = Worker(task)
            worker.signals.result.connect(self.print_output)
            worker.signals.finished.connect(self.thread_complete)
            thread_pool.start(worker)

if __name__ == '__main__':
    import sys
    app = QtWidgets.QApplication(sys.argv)
    window = MainWindow()
    window.show()
    sys.exit(app.exec_())