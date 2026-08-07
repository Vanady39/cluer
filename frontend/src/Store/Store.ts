import { configureStore } from "@reduxjs/toolkit";
import searchReduce from "../Reducers/searchReduce";

export const Store = configureStore({
  reducer: {
    search: searchReduce,
  },
});

export type RootState = ReturnType<typeof Store.getState>;
export type AppDispatch = typeof Store.dispatch;
