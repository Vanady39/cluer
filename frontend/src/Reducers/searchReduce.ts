import { createSlice, type PayloadAction } from "@reduxjs/toolkit";

interface SearchState {
  search: string;
}

const initialState: SearchState = {
  search: "",
};

const SearchSlice = createSlice({
  name: "search",
  initialState,
  reducers: {
    setSearch: (state, action: PayloadAction<string>) => {
      state.search = action.payload;
    },
    cleanSearch: (state) => {
      state.search = "";
    },
  },
});

export const { setSearch, cleanSearch } = SearchSlice.actions;
export default SearchSlice.reducer;
