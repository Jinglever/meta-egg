# 日志(log)

在 `pkg/log` 封装了一个日志组件，提供了常规的接口，如： 

- SetLevel
- Debug/Info/Warn/Error/Fatal
- Debugf/Infof/Warnf/Errorf/Fatalf
- WithFields/WithField/WithError

该组件的背后是用到了logrus日志组件。你可以自由地替换成其它日志组件，只需要实现了上述接口，就不会对工程里引用了log的地方产生任何破坏。