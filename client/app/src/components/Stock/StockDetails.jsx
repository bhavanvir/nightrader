import { useLocation } from "react-router-dom";

import BuyMarket from "./BuyMarket";

export default function StockDetails() {
  const state = useLocation();
  const { stock_id, stock_name, current_price } = state.state.stock;
  const Stock = {
    StockId: stock_id,
    StockName: stock_name,
    CurrentPrice: current_price,
  };

  return (
    <div>
      <div className="flex justify-center items-center font-bold text-4xl h-24 ">
        {Stock.StockName}
      </div>

      <div className="container mx-auto">
        <div className="grid grid-cols-2 gap-6">
          <BuyMarket Stock={Stock} />
        </div>
      </div>
    </div>
  );
}
