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

push "${cwd}" >/dev/null 2>&1

#. "${HOME}/.zshrc"

bsky_acount="${BSKY_ACCOUNT}"
bsky_apppass="${BSKY_APPPASS}"

temp_dir="./tmp"
if [[ ! -d "${temp_dir}" ]]; then
  mkdir -p "${temp_dir}"
fi
result_dir="./result"
if [[ ! -d "${result_dir}" ]]; then
  mkdir -p "${result_dir}"
fi

# Bingo ゲームのページのベースURL
bing_base_uri="https://bingo.b35.jp"
# Bingo ゲームのランキングページ
current_ranking_page="${bing_base_uri}/view_ranking.php"

# 日付文字列の取得
today=$(date "+%Y-%m-%d")
yesterday=$(date -v-1d "+%Y-%m-%d")

echo "BSKY_ACCOUNT: ${bsky_acount}"
echo "BSKY_APPPASS: ${bsky_apppass}"

if [[ -z "${bsky_acount}" ]]; then
  echo "BSKY_ACCOUNT is not set."
  exit 1
fi
if [[ -z "${bsky_apppass}" ]]; then
  echo "BSKY_APPPASS is not set."
  exit 1
fi

# 一時ファイルの削除
find "${temp_dir}" -print | while read line; do
  if [[ "${line}" == "${temp_dir}" ]]; then
    continue
  fi
  if [[ "${line}" == "${temp_dir}/.gitignore" ]]; then
    continue
  fi
  if [[ -f "${line}" ]]; then
    \rm -f "${line}"
  fi
done


# 昨日のランキング取得
current_ranking_file="${temp_dir}/bingo_ranking.html"
curl -sSL "${current_ranking_page}" -o "${current_ranking_file}"
if [[ $? -ne 0 ]]; then
  echo "curl bingo_ranking failed."
  exit 1
fi

# ランキングのhtmlからランキングのtsvを生成
read -d '' scriptVariable << 'EOF'
BEGIN{
  ramk=0;
  user_handle_url="";
  user_handle="";
  user_name="";
  user_point=0;
  user_history_uri="";
  # header
  printf "date\\trank\\tuser_handle\\tuser_name\\tuser_point\\tuser_history_uri\\n";
}
{
  # print $0
  match($0, /<td class="rank(| p[0-9]*)">([0-9]+)<\\/td>/, r);
  # print "a ", length(r)
  # for (i in r) {
    # print "a ", i, r[i]
  # }
  if (length(r) >= 1) {
    rank=r[2];
    user_handle_url="";
    user_handle="";
    user_name="";
    user_point=0;
    user_history_uri="";
  }
  # match($0, /\\<td class="user">\\<a href="([^"]+)"(| title="(.*)")>(.*)\\<\\/td>/, r)
  match($0, /<td class="user"><a href="([^"]+)">(.*)<\\/td>/, r)
  # print "b ", length(r);
  # for (i in r) {
  #   print "b ", i, r[i]
  # }
  if (length(r) >= 1) {
    user_handle_url=r[1];
    user_handle=r[2];
    user_name=r[2];
  }
  match($0, /<td class="user"><a href="([^"]+)" title="(.*)">(.*)<\\/td>/, r)
  # print "b ", length(r);
  # for (i in r) {
  #   print "b ", i, r[i]
  # }
  if (length(r) >= 1) {
    user_handle_url=r[1];
    user_handle=r[2];
    user_name=r[3];
  }
  match($0, /<td class="point">([0-9\\.]+)<\\/td>/, r);
  # print "c ", length(r)
  if (length(r) >= 1) {
    user_point=r[1];
  }
  
  match($0, /<td class="history"><a href="(.+)">履歴<\\/td>/, r);
  # print "d ", length(r)
  if (length(r) >= 1) {
    user_history_uri=r[1];
  }
  match($0, /<\\/tr>/, r);
  # print "z ", length(r)
  if (length(r) >= 1) {
    if (rank > 0) {
      printf "%s\\t%s\\t%s\\t%s\\t%s\\n", rank, user_handle, user_name, user_point, user_history_uri;
      rank=0;
      user_handle_url="";
      user_handle="";
      user_name="";
      user_point=0;
      user_history_uri="";
    }
  }
}
EOF

bingo_ranking_tsv="${temp_dir}/bingo_ranking.tsv"
gawk "${scriptVariable}" "${current_ranking_file}" > "${bingo_ranking_tsv}"
# 詳細のuriにベースを追加
bingo_ranking_uri_file="${temp_dir}/bingo_ranking_uri.tsv"
gawk -F'\t' 'BEGIN{OFS="\t"; l=0}{if(l==0){print $0;l=l+1;}else{printf "%s\t%s\t%s\t%s\t%s\t%s\n","'${yesterday}'",$1,$2,$3,$4,"'${bing_base_uri}'"$5; l=l+1;}}' "${temp_dir}/bingo_ranking.tsv" > "${bingo_ranking_uri_file}"

# Bluesky から、昨日の結果のポスト検索
# Bluesky にログイン
bingo_account_handle="bingo.b35.jp"
bsky login "${bsky_acount}" "${bsky_apppass}"
if [[ $? -ne 0 ]]; then
  echo "bsky login failed."
  exit 1
fi

# Bingo ゲームアカウントのタイムライン取得
bingo_account_tl="${temp_dir}/bingo_account_tl.json"
bingo_result_post_json="${temp_dir}/bingo_result_post.json"
bingo_result_post_json_tmp="${bingo_result_post_json}.tmp"
bsky tl -H "${bingo_account_handle}" -n 100 -json > "${bingo_account_tl}"

# タイムライン情報から、昨日の結果のポストを抽出
cat "${bingo_account_tl}" | grep -E '"createdAt":"'"${yesterday}" | grep -E '\[BINGO game result]' | head -1 > "${bingo_result_post_json_tmp}"
if [[ -z "$(cat "${bingo_result_post_json_tmp}")" ]]; then
  echo "bingo_result_post_json is empty."
  exit 1
fi
cat "${bingo_result_post_json_tmp}" | jq . > "${bingo_result_post_json}"
\rm -f "${bingo_result_post_json_tmp}"

# ポスト情報から ポストuri を抽出
post_uri_file="${temp_dir}/bingo_result_post_uri.txt"
cat "${bingo_result_post_json}" | jq -r .post.uri > "${post_uri_file}"
if [[ -z "$(cat "${post_uri_file}")" ]]; then
  echo "post_uri_file is empty."
  exit 1
fi

# ポストuriからポストデータを取得
result_post_file="${bingo_result_post_json}"
result_post_file_tmp="${result_post_file}.tmp"
bsky thread --json "$(cat "${post_uri_file}")" > "${result_post_file_tmp}"
if [[ -z "$(cat "${result_post_file}")" ]]; then
  echo "result_post_file is empty."
  \rm -f "${result_post_file_tmp}"
  exit 1
fi
  \rm -f "${result_post_file_tmp}"


# ポストから本文を抽出
post_text_file="${temp_dir}/bingo_result_post.txt"
cat "${bingo_result_post_json}" | jq -r .post.record.text > "${post_text_file}"

# # ポストから画像をダウンロード
post_image_file_uri="${temp_dir}/bingo_result_image_uri.txt"
post_image_file_path="${temp_dir}/bingo_result_image"
post_image_file_path_ext="jpg"

cat "${bingo_result_post_json}" | jq -r '.post.embed.images[0].fullsize' > "${post_image_file_uri}"
if [[ -z "$(cat "${post_image_file_uri}")" ]]; then
  echo "result_post_image_file_uri is empty."
  exit 1
fi
if [[ -n "$(cat "${post_image_file_uri}" | grep -E '@(jpeg|jpg)$')" ]]; then
  result_post_image_file_path_ext="jpg"
elif [[ -n "$(cat "${post_image_file_uri}" | grep -E '@(png)$')" ]]; then
  result_post_image_file_path_ext="png"
fi
post_image_file_path_file="${post_image_file_path}.${post_image_file_path_ext}"

curl -sSL "$(cat "${post_image_file_uri}")" -o "${post_image_file_path_file}"

# 取得したファイルから、必要なファイルを、日付を付けてコピー
result_bingo_ranking_url_tsv="${result_dir}/${yesterday}_bingo_ranking_url.tsv"
\mv "${bingo_ranking_uri_file}" "${result_bingo_ranking_url_tsv}"
result_current_ranking_html_file="${result_dir}/${yesterday}_bingo_ranking.html"
\mv "${current_ranking_file}" "${result_current_ranking_html_file}"
result_bingo_result_post_json_file="${result_dir}/${yesterday}_bingo_result_post.json"
\mv "${bingo_result_post_json}" "${result_bingo_result_post_json_file}"
result_post_uri_file="${result_dir}/${yesterday}_bingo_result_post_uri.txt"
\mv "${post_uri_file}" "${result_post_uri_file}"
result_post_image_file_path_file="${result_dir}/${yesterday}_bingo_result_image.${result_post_image_file_path_ext}"
\mv "${post_image_file_path_file}" "${result_post_image_file_path_file}"
result_post_text_file="${result_dir}/${yesterday}_bingo_result_post.txt"
\mv "${post_text_file}" "${result_post_text_file}"
pop >/dev/null 2>&1
