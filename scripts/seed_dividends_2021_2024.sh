#!/usr/bin/env bash
# Popula dividendos históricos 2021-2024 seguindo a mesma estrutura do seed 2025.
# Valores por pagamento derivados do histórico real (Investidor10) e distribuídos
# proporcionalmente pelos mesmos meses usados no seed de 2025.
#
# Uso: ./scripts/seed_dividends_2021_2024.sh [BASE_URL]
# Exemplo: ./scripts/seed_dividends_2021_2024.sh http://localhost:8080

set -euo pipefail

BASE_URL="${1:-http://localhost:8080}"
API="$BASE_URL/api/v1"

echo "==> Conectando em $API ..."

# --------------------------------------------------------------------------
# Busca IDs das ações cadastradas
# --------------------------------------------------------------------------
STOCKS_JSON=$(curl -sf "$API/stocks" || { echo "ERRO: API indisponível em $BASE_URL"; exit 1; })

get_id() {
  local ticker="$1"
  echo "$STOCKS_JSON" | grep -o "\"id\":[0-9]*,\"ticker\":\"$ticker\"" | grep -o '"id":[0-9]*' | grep -o '[0-9]*'
}

BBAS3_ID=$(get_id "BBAS3")
BBSE3_ID=$(get_id "BBSE3")
PETR4_ID=$(get_id "PETR4")
ITUB3_ID=$(get_id "ITUB3")
BRAP4_ID=$(get_id "BRAP4")
CMIG4_ID=$(get_id "CMIG4")
CPFE3_ID=$(get_id "CPFE3")
CSMG3_ID=$(get_id "CSMG3")
ISAE4_ID=$(get_id "ISAE4")
CXSE3_ID=$(get_id "CXSE3")

echo "IDs: BBAS3=$BBAS3_ID BBSE3=$BBSE3_ID PETR4=$PETR4_ID ITUB3=$ITUB3_ID BRAP4=$BRAP4_ID"
echo "     CMIG4=$CMIG4_ID CPFE3=$CPFE3_ID CSMG3=$CSMG3_ID ISAE4=$ISAE4_ID CXSE3=$CXSE3_ID"

TOTAL=0
ERROS=0

# --------------------------------------------------------------------------
# Função de inserção
# --------------------------------------------------------------------------
post_dividend() {
  local stock_id="$1" amount="$2" month="$3" year="$4" type="$5"
  local body
  body=$(printf '{"amount":%s,"month":%d,"year":%d,"type":"%s"}' "$amount" "$month" "$year" "$type")
  local http_code
  http_code=$(curl -sf -o /dev/null -w "%{http_code}" \
    -X POST "$API/stocks/$stock_id/dividends" \
    -H "Content-Type: application/json" \
    -d "$body" 2>&1) || true

  if [[ "$http_code" =~ ^2 ]]; then
    TOTAL=$((TOTAL + 1))
  else
    echo "  AVISO: stock_id=$stock_id month=$month/$year type=$type → HTTP $http_code"
    ERROS=$((ERROS + 1))
  fi
}

# --------------------------------------------------------------------------
# Dados históricos por ano
# Mesmos meses do seed 2025; amounts derivados do histórico real de cada ano.
#
# BBAS3  → meses [1,4,6,9,12]   tipo: jcp
# BBSE3  → meses [1,3,6,8,11]   tipo: dividendo
# PETR4  → meses [2,5,8,11]     tipo: dividendo
# ITUB3  → meses [1..12]        tipo: jcp
# BRAP4  → meses [3,6,9,12]     tipo: dividendo
# CMIG4  → meses [3,6,9,12]     tipo: dividendo
# CPFE3  → meses [2,5,8,11]     tipo: dividendo
# CSMG3  → meses [4,8,12]       tipo: jcp
# ISAE4  → meses [3,6,9,12]     tipo: jcp
# CXSE3  → meses [3,6,9,12]     tipo: dividendo
# --------------------------------------------------------------------------

seed_year() {
  local year="$1"
  local \
    bbas3_amt bbse3_amt petr4_amt itub3_amt brap4_amt \
    cmig4_amt cpfe3_amt csmg3_amt isae4_amt cxse3_amt

  case "$year" in
    2021)
      bbas3_amt=0.12; bbse3_amt=0.20; petr4_amt=0.60; itub3_amt=0.05; brap4_amt=1.96
      cmig4_amt=0.22; cpfe3_amt=0.75; csmg3_amt=0.25; isae4_amt=0.37; cxse3_amt=0.12
      ;;
    2022)
      bbas3_amt=0.21; bbse3_amt=0.39; petr4_amt=2.93; itub3_amt=0.05; brap4_amt=0.78
      cmig4_amt=0.19; cpfe3_amt=0.81; csmg3_amt=0.13; isae4_amt=0.27; cxse3_amt=0.16
      ;;
    2023)
      bbas3_amt=0.23; bbse3_amt=0.69; petr4_amt=1.53; itub3_amt=0.05; brap4_amt=0.80
      cmig4_amt=0.08; cpfe3_amt=0.53; csmg3_amt=0.56; isae4_amt=0.61; cxse3_amt=0.25
      ;;
    2024)
      bbas3_amt=0.27; bbse3_amt=0.53; petr4_amt=1.29; itub3_amt=0.12; brap4_amt=0.52
      cmig4_amt=0.08; cpfe3_amt=0.69; csmg3_amt=0.69; isae4_amt=0.55; cxse3_amt=0.27
      ;;
    *) echo "Ano desconhecido: $year"; return 1 ;;
  esac

  echo "--> Ano $year"

  # BBAS3 – JCP trimestral + extras (meses 1,4,6,9,12)
  for m in 1 4 6 9 12; do
    post_dividend "$BBAS3_ID" "$bbas3_amt" "$m" "$year" "jcp"
  done

  # BBSE3 – Dividendos semestrais distribuídos (meses 1,3,6,8,11)
  for m in 1 3 6 8 11; do
    post_dividend "$BBSE3_ID" "$bbse3_amt" "$m" "$year" "dividendo"
  done

  # PETR4 – Dividendos trimestrais (meses 2,5,8,11)
  for m in 2 5 8 11; do
    post_dividend "$PETR4_ID" "$petr4_amt" "$m" "$year" "dividendo"
  done

  # ITUB3 – JCP/Dividendos mensais (meses 1 a 12)
  for m in 1 2 3 4 5 6 7 8 9 10 11 12; do
    post_dividend "$ITUB3_ID" "$itub3_amt" "$m" "$year" "jcp"
  done

  # BRAP4 – Dividendos trimestrais (meses 3,6,9,12)
  for m in 3 6 9 12; do
    post_dividend "$BRAP4_ID" "$brap4_amt" "$m" "$year" "dividendo"
  done

  # CMIG4 – Dividendos semestrais (meses 3,6,9,12)
  for m in 3 6 9 12; do
    post_dividend "$CMIG4_ID" "$cmig4_amt" "$m" "$year" "dividendo"
  done

  # CPFE3 – Dividendos trimestrais (meses 2,5,8,11)
  for m in 2 5 8 11; do
    post_dividend "$CPFE3_ID" "$cpfe3_amt" "$m" "$year" "dividendo"
  done

  # CSMG3 – JCP trimestral (meses 4,8,12)
  for m in 4 8 12; do
    post_dividend "$CSMG3_ID" "$csmg3_amt" "$m" "$year" "jcp"
  done

  # ISAE4 – JCP trimestral (meses 3,6,9,12)
  for m in 3 6 9 12; do
    post_dividend "$ISAE4_ID" "$isae4_amt" "$m" "$year" "jcp"
  done

  # CXSE3 – Dividendos trimestrais (meses 3,6,9,12)
  for m in 3 6 9 12; do
    post_dividend "$CXSE3_ID" "$cxse3_amt" "$m" "$year" "dividendo"
  done
}

# --------------------------------------------------------------------------
# Executa para cada ano
# --------------------------------------------------------------------------
for year in 2021 2022 2023 2024; do
  seed_year "$year"
done

echo ""
echo "==> Concluído: $TOTAL dividendos inseridos, $ERROS erros."
