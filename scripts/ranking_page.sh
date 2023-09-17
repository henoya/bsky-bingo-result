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

start_date="202309"
today=$(date "+%Y%m")
end_date="${today}"
cur_date="${start_date}"
while [[ "${cur_date}" -le "${end_date}" ]]; do
  echo "cur_date: ${cur_date}"
  curl -H "Cookie: ${cookie}" "https://bingo.b35.jp/view_ranking.php?m=${cur_date}" > "./result/${cur_date}.html"

  cat "./result/${cur_date}.html" | grep -E '<td class="history.+履歴' | sed -E 's|(<td.+a href=")(view[^"]+)">履歴</td>|https://bingo.b35.jp/\2|' >"./result/${cur_date}_month_ranking.txt"
  cat "./result/${cur_date}_month_ranking.txt" | while read line; do
    echo "line: ${line}"
    did=$(echo "${line}" | sed -E 's|(.+)(did%3Aplc%3A.+)|\2|')
    echo "did: ${did}"
    curl -H "Cookie: ${cookie}" "${line}" > "./result/${cur_date}/${did}_detail.html"
  done
  cur_date=$(date -v+1m -j -f "%Y%m" "${cur_date}" "+%Y%m")
done

pop >/dev/null 2>&1
