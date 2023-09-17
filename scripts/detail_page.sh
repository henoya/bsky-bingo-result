#!/bin/zsh
# ${0} の dirname を取得
cwd=$(dirname "${0}")

# ${0} が 相対パスの場合は cd して pwd を取得
expr "${0}" : "/.*" > /dev/null || cwd=$( (cd "${cwd}" && pwd) )

# CMD_PATH        スクリプトのパスを取得(スクリプト名込み)
CMD_PATH="${cwd}"
# CMD_NAME        スクリプトのパスから、ファイル名部分を取得
CMD_NAME=${CMD_PATH##*/}
# CMD_BASE_NAME   スクリプトのファイル名からベース文字列を取得
CMD_BASE_NAME="${CMD_PATH}"

work_dir="$(realpath "${cwd}/..")"
bin_dir="${work_dir}/bin"
temp_dir="${work_dir}/tmp"

push "${work_dir}" >/dev/null 2>&1

cookie="uh=Kw6ySfyUdfGbqMzDlS3elw%3D%3D; up=LEl7gpUipO4C%2B735IVolAKw2RCl8eH5%2Ftn%2BMCelKCt4%3D; PHPSESSID=umh9dvbe0j51gto03ib6u1ii7f"

start_date="20230901"
today=$(date "+%Y%m%d")
end_date="${today}"
cur_date="${start_date}"
while [[ "${cur_date}" -lt "${end_date}" ]]; do
  target_month="$(date -j -f "%Y%m%d" "${cur_date}" "+%Y%m")"
  target_dir="${work_dir}/result/${target_month}/day_detail"
  target_file="${target_dir}/${cur_date}.html"
  if [[ ! -f "${target_file}" ]]; then
    echo "cur_date: ${cur_date}"
    curl -H "Cookie: ${cookie}" "https://bingo.b35.jp/result.php?d=${cur_date}" >  "${target_file}"
  fi
  cur_date=$(date -v+1d -j -f "%Y%m%d" "${cur_date}" "+%Y%m%d")
done

pop >/dev/null 2>&1
