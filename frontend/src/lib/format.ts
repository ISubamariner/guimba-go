import type { MoneyDTO } from "@/types/api";

export function formatMoney(money: MoneyDTO): string {
  const amount = parseFloat(money.amount);
  return `${money.currency} ${amount.toLocaleString(undefined, {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  })}`;
}

export function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString();
}

export function formatDateTime(dateStr: string): string {
  return new Date(dateStr).toLocaleString();
}
