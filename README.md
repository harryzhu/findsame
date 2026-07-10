# findsame

find same(duplicate) files in a folder quickly.

## Usage

```Bash
./findsame --source-dir="/Volumes/HDD4/v2"

# --log-dir= : the path for saving result, default: ./logs
# --debug=true|false : if show debug info, default: false
# --serial=true|false : if your device is HDD(not SSD), pls set --serial=true
#

# 机械硬盘需要设置参数 --serial
#
./findsame --source-dir="/Volumes/HDD4/v2" --serial
#
```

## Result

在 `--log-dir=` 指定的路径下，有 `empty-files.html` 和 `same-files.html` 两个文件.

1) `empty-files.html`: 所有空文件的路径

2) `same-files.html`: 内容完全相同（重复）的文件，按组分开列表


