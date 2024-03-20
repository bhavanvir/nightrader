import { useLocation } from "react-router-dom";

export default function StockDetails() {
  const state = useLocation();
  const { stock_id, stock_name, current_price } = state.state.stock;

  return (
    <div>
      <h1>{stock_name}</h1>
    </div>
  );
}
