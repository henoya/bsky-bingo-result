#!/bin/bash

cookie="uh=Kw6ySfyUdfGbqMzDlS3elw%3D%3D; up=LEl7gpUipO4C%2B735IVolAKw2RCl8eH5%2Ftn%2BMCelKCt4%3D; PHPSESSID=umh9dvbe0j51gto03ib6u1ii7f"

start_date="202309"
cur_date="202309"
cur_month=$(date -v-1d "+%Y-%m")
cat "result/${cur_date}/${cur_month}_bingo_ranking.tsv" | sed -E 's|^([0-9-]+)\t([0-9]*)\t(.+)\t(.+)\t(.+)\t(http.+did=)(.+)|\1\t\2\t\3\t\4\t\7|' > "result/${cur_date}/${cur_month}_bingo_ranking_list.tsv"
cat "result/${cur_date}/${cur_month}_bingo_ranking.tsv" | sed -E 's|^([0-9-]+)\t([0-9]*)\t(.+)\t(.+)\t(.+)\t(http.+did=)(.+)|\7|' > "result/${cur_date}/${cur_month}_bingo_did_list.tsv"

month_dir="result/${cur_date}"
month_mvp="${month_dir}/${cur_month}_bingo_mvp.tsv"
printf "did\tsanka\tteam\tbingo\tsolo\n" > "${month_mvp}"
cat "${month_dir}/${cur_month}_bingo_did_list.tsv" | while read line; do
  did="${line}"
  detail_file="${month_dir}/${did}_detail.html"
  if [[ -f "${detail_file}" ]]; then
    printf "${did}\t" >> "${month_mvp}"
    sanka="$(cat "${detail_file}" | grep -E '<td class="category">参加</td>' | wc -l | sed -E 's| +||g')"
    team="$(cat "${detail_file}" | grep -E '<td class="category">チーム勝利</td>' | wc -l | sed -E 's| +||g')"
    bingo="$(cat "${detail_file}" | grep -E '<td class="category">ビンゴ達成</td>' | wc -l | sed -E 's| +||g')"
    solo="$(cat "${detail_file}" | grep -E '<td class="category">単独ボーナス</td>' | wc -l | sed -E 's| +||g')"
    printf "${sanka}\t${team}\t${bingo}\t${solo}\n" >> "${month_mvp}"
  fi
done