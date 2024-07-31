// Generated by github.com/adamluo159/tabtoy
// Version: 
// DO NOT EDIT!!
using System.Collections.Generic;

namespace table
{
	
	// Defined in table: Globals
	public enum ActorType
	{
		
		
		Leader = 0, // 唐僧
		
		
		Monkey = 1, // 孙悟空
		
		
		Pig = 2, // 猪八戒
		
		
		Hammer = 3, // 沙僧
	
	}
	
	

	// Defined in table: Config
	
	public partial class Config
	{
	
		public tabtoy.Logger TableLogger = new tabtoy.Logger();
	
		
		/// <summary> 
		/// AAA
		/// </summary>
		public List<AAADefine> AAA = new List<AAADefine>(); 
	
	
		#region Index code
	 	Dictionary<int, AAADefine> _AAAByID = new Dictionary<int, AAADefine>();
        public AAADefine GetAAAByID(int ID, AAADefine def = default(AAADefine))
        {
            AAADefine ret;
            if ( _AAAByID.TryGetValue( ID, out ret ) )
            {
                return ret;
            }
			
			if ( def == default(AAADefine) )
			{
				TableLogger.ErrorLine("GetAAAByID failed, ID: {0}", ID);
			}

            return def;
        }
		
		public string GetBuildID(){
			return "a48111a987b48412d022473d559baff4";
		}
	
		#endregion
		#region Deserialize code
		
		static tabtoy.DeserializeHandler<Config> _ConfigDeserializeHandler;
		static tabtoy.DeserializeHandler<Config> ConfigDeserializeHandler
		{
			get
			{
				if (_ConfigDeserializeHandler == null )
				{
					_ConfigDeserializeHandler = new tabtoy.DeserializeHandler<Config>(Deserialize);
				}

				return _ConfigDeserializeHandler;
			}
		}
		public static void Deserialize( Config ins, tabtoy.DataReader reader )
		{
			
 			int tag = -1;
            while ( -1 != (tag = reader.ReadTag()))
            {
                switch (tag)
                { 
                	case 0xa0000:
                	{
						ins.AAA.Add( reader.ReadStruct<AAADefine>(AAADefineDeserializeHandler) );
                	}
                	break; 
                }
             } 

			
			// Build AAA Index
			for( int i = 0;i< ins.AAA.Count;i++)
			{
				var element = ins.AAA[i];
				
				ins._AAAByID.Add(element.ID, element);
				
			}
			
		}
		static tabtoy.DeserializeHandler<Vec2> _Vec2DeserializeHandler;
		static tabtoy.DeserializeHandler<Vec2> Vec2DeserializeHandler
		{
			get
			{
				if (_Vec2DeserializeHandler == null )
				{
					_Vec2DeserializeHandler = new tabtoy.DeserializeHandler<Vec2>(Deserialize);
				}

				return _Vec2DeserializeHandler;
			}
		}
		public static void Deserialize( Vec2 ins, tabtoy.DataReader reader )
		{
			
 			int tag = -1;
            while ( -1 != (tag = reader.ReadTag()))
            {
                switch (tag)
                { 
                	case 0x10000:
                	{
						ins.X = reader.ReadInt32();
                	}
                	break; 
                	case 0x10001:
                	{
						ins.Y = reader.ReadInt32();
                	}
                	break; 
                }
             } 

			
		}
		static tabtoy.DeserializeHandler<AAADefine> _AAADefineDeserializeHandler;
		static tabtoy.DeserializeHandler<AAADefine> AAADefineDeserializeHandler
		{
			get
			{
				if (_AAADefineDeserializeHandler == null )
				{
					_AAADefineDeserializeHandler = new tabtoy.DeserializeHandler<AAADefine>(Deserialize);
				}

				return _AAADefineDeserializeHandler;
			}
		}
		public static void Deserialize( AAADefine ins, tabtoy.DataReader reader )
		{
			
 			int tag = -1;
            while ( -1 != (tag = reader.ReadTag()))
            {
                switch (tag)
                { 
                	case 0x10000:
                	{
						ins.ID = reader.ReadInt32();
                	}
                	break; 
                	case 0x60001:
                	{
						ins.Name = reader.ReadString();
                	}
                	break; 
                	case 0x60002:
                	{
						ins.SSS = reader.ReadString();
                	}
                	break; 
                	case 0x90003:
                	{
						ins.DDD = reader.ReadStruct<Vec2>(Vec2DeserializeHandler);
                	}
                	break; 
                }
             } 

			
		}
		#endregion
		#region Clear Code
		public void Clear( )
		{			
				AAA.Clear(); 
			
				_AAAByID.Clear(); 
		}
		#endregion
	

	} 

	// Defined in table: Globals
	
	public partial class Vec2
	{
	
		
		
		public int X = 0; 
		
		
		public int Y = 0; 
	
	

	} 

	// Defined in table: AAA
	[System.Serializable]
	public partial class AAADefine
	{
	
		
		/// <summary> 
		/// 唯一ID
		/// </summary>
		public int ID = 0; 
		
		/// <summary> 
		/// 名称
		/// </summary>
		public string Name = ""; 
		
		/// <summary> 
		/// 名称
		/// </summary>
		public string SSS = ""; 
		
		
		public Vec2 DDD = new Vec2(); 
	
	

	} 

}
