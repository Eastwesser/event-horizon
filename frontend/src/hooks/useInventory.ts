import { useEffect } from 'react';
import { useInventoryStore } from '../store/inventoryStore';

export const useInventory = () => {
  const {
    items,
    total,
    loading,
    error,
    selectedItem,
    fetchItems,
    fetchItem,
    createItem,
    updateItem,
    deleteItem,
    clearSelected,
    clearError,
  } = useInventoryStore();

  useEffect(() => {
    // Автоматическая очистка ошибки через 5 секунд
    if (error) {
      const timer = setTimeout(() => clearError(), 5000);
      return () => clearTimeout(timer);
    }
  }, [error, clearError]);

  return {
    items,
    total,
    loading,
    error,
    selectedItem,
    fetchItems,
    fetchItem,
    createItem,
    updateItem,
    deleteItem,
    clearSelected,
    clearError,
  };
};
