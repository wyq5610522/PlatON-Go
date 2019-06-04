#!/bin/sh
set -e


# ubuntu builder
function ubuntu_build() {
    cd ${PLATON_PATH}
    [ "${BUILD_WITH_MV}" == "true" ] && build_command="make all-with-mv"
    [ "${BUILD_WITH_MV}" == "fales" ] && build_command="make all"
    chmod u+x ./build/*.sh
    ${build_command}
}

# ubuntu package
function ubuntu_package() {
    cd ${PLATON_PATH}/build/bin/
    [ "${BUILD_WITH_MV}" == "true" ] && postfix="-with-mv"
    [ "${BUILD_WITH_MV}" == "fales" ] && postfix=""
	dir_name="platon-ubuntu${postfix}"
	mkdir ${dir_name} && cp platon ctool ethkey ${dir_name}
	tar_name="${dir_name}.tar.gz"
	tar -zcf ${tar_name} ${dir_name}
}

# windows builder
function windows_build() {
    cd ${PLATON_PATH}
    [ "${BUILD_WITH_MV}" == "true" ] && build_command="go run build/ci.go install -mpc on"
    [ "${BUILD_WITH_MV}" == "fales" ] && build_command="go run build/ci.go install"
    chmod u+x ./build/*.sh
    sh ./build/build_deps.sh
    ${build_command}
}

# windows package
function windows_package() {
    cd ${PLATON_PATH}/build/bin/
    [ "${BUILD_WITH_MV}" == "true" ] && postfix="-with-mv"
    [ "${BUILD_WITH_MV}" == "fales" ] && postfix=""
	dir_name="platon-windows${postfix}"
	mkdir ${dir_name} && cp platon.exe ctool.exe ethkey.exe ${dir_name}
	zip_name="${dir_name}.zip"
	zip -r ${zip_name} ${dir_name} 1>$-
}

# upload packages
function upload_package() {
    cd ${PLATON_PATH}
    remote_ip="58.250.250.235"
    remote_port="3122"
    remote_username="platon"
    remote_password="${HTTP_SERVER_PASSWORD}"
    remote_path="/home/platon/Jenkins/ci/travis/packages/${TRAVIS_BUILD_NUMBER}"
    local_path="${PLATON_PATH}/build/bin"
    file_name=$1
    python push_file_to_remote_linux.py "${remote_ip}" "${remote_username}" "${remote_password}" "${remote_path}/${file_name}" "${local_path}/${file_name}"
}


# main
# Receive parameters
if [ ! -f "./build/ci.sh" ]; then
    echo "$0 must be run from the root of the platon repository."
    exit 2
elif [ $# != 3 ]; then 
    echo "Usage:"
    echo " ci.sh [platon_root_path] [current_system] [is_open_mv]"
    exit 2
else
    echo "PRARMS:"
    cd $1 && PLATON_PATH=$(pwd) && echo "PLATON_PATH=${PLATON_PATH}"
    BUILD_PLATFORM=$2 && echo "BUILD_PLATFORM=${BUILD_PLATFORM}"
    BUILD_WITH_MV=$3 && echo "BUILD_WITH_MV=${BUILD_WITH_MV}"
fi

# run build
cd ${PLATON_PATH}
if [ "${BUILD_PLATFORM}" == "ubuntu" ]; then
    ubuntu_build && ubuntu_package && upload_package ${tar_name}
else [ "${BUILD_PLATFORM}" == "windows" ]
    windows_build && windows_package && upload_package ${zip_name}
fi

# exit
exit 0